package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

type Manager struct {
	Store         *Store
	Engine        media.Engine
	Root          string
	Workers       int
	MaxBytes      int64
	Timeout       time.Duration
	StorageBudget int64
	storageMu     sync.Mutex
	reservedBytes int64
	mu            sync.Mutex
	active        map[string]context.CancelFunc
	wg            sync.WaitGroup
	wake          chan struct{}
}

func (m *Manager) Start(ctx context.Context) error {
	if err := os.MkdirAll(m.Root, 0700); err != nil {
		return err
	}
	if err := m.Store.Recover(); err != nil {
		return err
	}
	m.active = map[string]context.CancelFunc{}
	m.wake = make(chan struct{}, 1)
	// Every non-complete directory is scratch, including work interrupted at restart.
	entries, err := os.ReadDir(m.Root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !validJobID(entry.Name()) {
			continue
		}
		j, err := scan(m.Store.db.QueryRow(`SELECT body,url,option_id,owner FROM jobs WHERE id=?`, entry.Name()))
		if err != nil || j.State != "complete" {
			_ = m.remove(entry.Name())
		}
	}
	for i := 0; i < m.Workers; i++ {
		m.wg.Add(1)
		go m.worker(ctx)
	}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		m.sweep()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.sweep()
			}
		}
	}()
	return nil
}
func (m *Manager) Wait() { m.wg.Wait() }
func (m *Manager) Notify() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}
func (m *Manager) worker(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		j, err := m.Store.Claim()
		if err == nil {
			m.run(ctx, j)
			continue
		}
		if !errors.Is(err, ErrNotFound) {
			slog.Error("queue claim failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-m.wake:
		}
	}
}
func (m *Manager) run(parent context.Context, j Job) {
	release, err := m.reserveStorage()
	if err != nil {
		m.fail(j, "Temporary storage is full. Try again after older files expire.")
		return
	}
	defer release()
	ctx, cancel := context.WithTimeout(parent, m.Timeout)
	defer cancel()
	m.mu.Lock()
	m.active[j.ID] = cancel
	m.mu.Unlock()
	defer func() { m.mu.Lock(); delete(m.active, j.ID); m.mu.Unlock() }()
	current, err := m.Store.Get(j.ID, j.Owner)
	if err != nil || current.State != "processing" {
		return
	}
	dir := filepath.Join(m.Root, j.ID)
	if err = os.Mkdir(dir, 0700); err != nil {
		m.fail(j, "Temporary storage could not be created.")
		return
	}
	var watchdog sync.WaitGroup
	watchdog.Add(1)
	limitExceeded := make(chan struct{}, 1)
	go func() {
		defer watchdog.Done()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var total int64
				err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.Type()&os.ModeSymlink != 0 {
						return errors.New("symlink")
					}
					if !d.IsDir() {
						info, err := d.Info()
						if err != nil {
							return err
						}
						total += info.Size()
					}
					return nil
				})
				if err != nil || total > m.MaxBytes {
					select {
					case limitExceeded <- struct{}{}:
					default:
					}
					cancel()
					return
				}
			}
		}
	}()
	analysis, err := m.Engine.Analyze(ctx, j.URL)
	var option media.Option
	found := false
	if err == nil {
		for _, candidate := range analysis.Options {
			if candidate.ID == j.OptionID {
				option = candidate
				found = true
				break
			}
		}
		if !found {
			err = errors.New("This format is no longer available. Analyze the link again.")
		}
	}
	var outputs []media.Output
	last := time.Time{}
	if err == nil {
		outputs, err = m.Engine.Download(ctx, analysis, option, dir, func(p media.Progress) {
			if time.Since(last) < 350*time.Millisecond {
				return
			}
			last = time.Now()
			_, _ = m.Store.Transition(j.ID, j.Owner, []string{"processing"}, func(j *Job) { j.Phase = p.Phase; j.Percent = p.Percent })
		})
	}
	cancel()
	watchdog.Wait()
	select {
	case <-limitExceeded:
		err = errors.New("Processing exceeded the temporary storage limit.")
	default:
	}
	if err == nil {
		files := []File{}
		var total int64
		for i, output := range outputs {
			rel, relErr := filepath.Rel(dir, output.Path)
			info, statErr := os.Lstat(output.Path)
			if relErr != nil || strings.Contains(rel, string(filepath.Separator)) || rel == ".." || info == nil || statErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
				err = errors.New("Unsafe processing output was rejected.")
				break
			}
			total += info.Size()
			files = append(files, File{ID: fmt.Sprint(i), Name: rel, Bytes: info.Size()})
		}
		if total > m.MaxBytes || len(files) == 0 {
			err = errors.New("Output exceeds the file limit or contains no downloadable files.")
		}
		if err == nil {
			_, err = m.Store.Transition(j.ID, j.Owner, []string{"processing"}, func(j *Job) { j.State = "complete"; j.Phase = "complete"; j.Percent = 100; j.Files = files })
		}
	}
	if err != nil {
		m.fail(j, err.Error())
		_ = m.remove(j.ID)
	}
}
func (m *Manager) fail(j Job, message string) {
	_, _ = m.Store.Transition(j.ID, j.Owner, []string{"processing"}, func(j *Job) { j.State = "failed"; j.Phase = "failed"; j.Error = message })
}
func (m *Manager) Cancel(id, owner string) (Job, error) {
	j, err := m.Store.Transition(id, owner, []string{"queued", "processing"}, func(j *Job) { j.State = "canceled"; j.Phase = "canceled"; j.Error = "" })
	if err != nil {
		return j, err
	}
	m.mu.Lock()
	if cancel := m.active[id]; cancel != nil {
		cancel()
	}
	m.mu.Unlock()
	return j, nil
}
func validJobID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			if r < 'a' || r > 'f' {
				return false
			}
		}
	}
	return true
}
func (m *Manager) remove(id string) error {
	if !validJobID(id) {
		return errors.New("unsafe directory id")
	}
	target := filepath.Join(m.Root, id)
	rel, err := filepath.Rel(m.Root, target)
	if err != nil || rel != id {
		return errors.New("unsafe directory")
	}
	return os.RemoveAll(target)
}
func (m *Manager) Delete(id, owner string) error {
	m.mu.Lock()
	_, active := m.active[id]
	m.mu.Unlock()
	if active {
		return errors.New("cancellation is finishing; try again in a moment")
	}
	if err := m.Store.Delete(id, owner); err != nil {
		return err
	}
	return m.remove(id)
}
func (m *Manager) sweep() {
	expired, err := m.Store.Expired(time.Now())
	if err != nil {
		slog.Error("expiry scan failed", "error", err)
		return
	}
	for _, j := range expired {
		if j.State == "queued" || j.State == "processing" {
			_, _ = m.Cancel(j.ID, j.Owner)
		}
		if err := m.Delete(j.ID, j.Owner); err != nil {
			slog.Debug("expiry deferred", "job_id", j.ID)
		}
	}
}
