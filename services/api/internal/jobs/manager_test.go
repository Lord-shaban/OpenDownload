package jobs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

type controlledEngine struct {
	oversize bool
	started  chan struct{}
	finished chan struct{}
}

func (e controlledEngine) Analyze(context.Context, string) (media.Analysis, error) {
	return media.Analysis{Options: []media.Option{{ID: "video"}}}, nil
}
func (e controlledEngine) Download(ctx context.Context, a media.Analysis, o media.Option, dir string, update func(media.Progress)) ([]media.Output, error) {
	defer close(e.finished)
	data := []byte("partial")
	if e.oversize {
		data = make([]byte, 4096)
	}
	if err := os.WriteFile(filepath.Join(dir, "media.mp4"), data, 0600); err != nil {
		return nil, err
	}
	close(e.started)
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestWorkerCancellationAndStorageWatchdog(t *testing.T) {
	for _, oversize := range []bool{false, true} {
		name := "cancel"
		if oversize {
			name = "storage-limit"
		}
		t.Run(name, func(t *testing.T) {
			s := testStore(t, 5)
			engine := controlledEngine{oversize: oversize, started: make(chan struct{}), finished: make(chan struct{})}
			m := &Manager{Store: s, Engine: engine, Root: filepath.Join(t.TempDir(), "files"), Workers: 1, MaxBytes: 1024, Timeout: 3 * time.Second}
			ctx, stop := context.WithCancel(context.Background())
			if err := m.Start(ctx); err != nil {
				t.Fatal(err)
			}
			defer func() { stop(); m.Wait() }()
			j, err := s.Create(sample("owner"))
			if err != nil {
				t.Fatal(err)
			}
			m.Notify()
			select {
			case <-engine.started:
			case <-time.After(2 * time.Second):
				t.Fatal("worker did not start")
			}
			if !oversize {
				if _, err := m.Cancel(j.ID, j.Owner); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-engine.finished:
			case <-time.After(2 * time.Second):
				t.Fatal("worker was not stopped")
			}
			// Wait for the worker to persist its terminal state and remove scratch files.
			deadline := time.Now().Add(time.Second)
			want := "canceled"
			if oversize {
				want = "failed"
			}
			for {
				current, err := s.Get(j.ID, j.Owner)
				if err != nil {
					t.Fatal(err)
				}
				_, statErr := os.Stat(filepath.Join(m.Root, j.ID))
				if current.State == want && os.IsNotExist(statErr) {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("state=%s scratch=%v", current.State, statErr)
				}
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
func TestSweepRemovesExpiredFilesAndDatabaseRow(t *testing.T) {
	s := testStore(t, 5)
	j := sample("owner")
	j.ExpiresAt = time.Now().Add(-time.Minute)
	j, err := s.Create(j)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Transition(j.ID, j.Owner, []string{"processing"}, func(j *Job) { j.State = "complete" }); err != nil {
		t.Fatal(err)
	}
	m := &Manager{Store: s, Root: t.TempDir(), active: map[string]context.CancelFunc{}}
	dir := filepath.Join(m.Root, j.ID)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "media.mp4"), []byte("sample"), 0600); err != nil {
		t.Fatal(err)
	}
	m.sweep()
	if _, err := s.Get(j.ID, j.Owner); err != ErrNotFound {
		t.Fatalf("expired record remains: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("expired files remain")
	}
}
