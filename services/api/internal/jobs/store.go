// Package jobs owns durable state transitions and the bounded background queue.
package jobs

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("job not found")
var ErrCapacity = errors.New("the queue is full; wait for a job to finish")
var ErrState = errors.New("this job cannot be changed in its current state")

type File struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
}
type Job struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Platform  string    `json:"platform"`
	Kind      string    `json:"kind"`
	Label     string    `json:"label"`
	Extension string    `json:"extension"`
	State     string    `json:"state"`
	Phase     string    `json:"phase"`
	Percent   float64   `json:"percent"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Files     []File    `json:"files"`
	URL       string    `json:"-"`
	OptionID  string    `json:"-"`
	Owner     string    `json:"-"`
}
type Store struct {
	db       *sql.DB
	mu       sync.Mutex
	Capacity int
}

func ID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func Open(path string, capacity int) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; CREATE TABLE IF NOT EXISTS jobs (id TEXT PRIMARY KEY, owner TEXT NOT NULL, state TEXT NOT NULL, created INTEGER NOT NULL, expires INTEGER NOT NULL, url TEXT NOT NULL, option_id TEXT NOT NULL, body BLOB NOT NULL); CREATE INDEX IF NOT EXISTS jobs_owner ON jobs(owner,created); CREATE INDEX IF NOT EXISTS jobs_queue ON jobs(state,created);`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db, Capacity: capacity}, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) save(j Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE jobs SET state=?,body=? WHERE id=?`, j.State, b, j.ID)
	return err
}
func scan(row interface{ Scan(...any) error }) (Job, error) {
	var j Job
	var body []byte
	err := row.Scan(&body, &j.URL, &j.OptionID, &j.Owner)
	if err == sql.ErrNoRows {
		return j, ErrNotFound
	}
	if err != nil {
		return j, err
	}
	err = json.Unmarshal(body, &j)
	return j, err
}

func (s *Store) Create(j Job) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var active, owned, total int
	if err := s.db.QueryRow(`SELECT count(*), coalesce(sum(owner=?),0) FROM jobs WHERE state IN ('queued','processing')`, j.Owner).Scan(&active, &owned); err != nil {
		return Job{}, err
	}
	if err := s.db.QueryRow(`SELECT count(*) FROM jobs`).Scan(&total); err != nil {
		return Job{}, err
	}
	if active >= s.Capacity || owned >= 5 || total >= 1000 {
		return Job{}, ErrCapacity
	}
	j.ID = ID()
	j.State = "queued"
	j.Phase = "queued"
	j.Percent = -1
	j.Files = []File{}
	b, err := json.Marshal(j)
	if err != nil {
		return Job{}, err
	}
	_, err = s.db.Exec(`INSERT INTO jobs (id,owner,state,created,expires,url,option_id,body) VALUES(?,?,?,?,?,?,?,?)`, j.ID, j.Owner, j.State, j.CreatedAt.UnixMilli(), j.ExpiresAt.Unix(), j.URL, j.OptionID, b)
	return j, err
}
func (s *Store) Get(id, owner string) (Job, error) {
	return scan(s.db.QueryRow(`SELECT body,url,option_id,owner FROM jobs WHERE id=? AND owner=?`, id, owner))
}
func (s *Store) List(owner string) ([]Job, error) {
	rows, err := s.db.Query(`SELECT body,url,option_id,owner FROM jobs WHERE owner=? ORDER BY created DESC LIMIT 100`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Job{}
	for rows.Next() {
		j, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (s *Store) Claim() (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, err := scan(s.db.QueryRow(`SELECT body,url,option_id,owner FROM jobs WHERE state='queued' ORDER BY created LIMIT 1`))
	if err != nil {
		return j, err
	}
	j.State = "processing"
	j.Phase = "analyzing"
	return j, s.save(j)
}
func (s *Store) Transition(id, owner string, allowed []string, change func(*Job)) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, err := s.Get(id, owner)
	if err != nil {
		return j, err
	}
	ok := false
	for _, state := range allowed {
		if state == j.State {
			ok = true
		}
	}
	if !ok {
		return j, ErrState
	}
	change(&j)
	return j, s.save(j)
}
func (s *Store) Delete(id, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, err := s.Get(id, owner)
	if err != nil {
		return err
	}
	if j.State == "processing" || j.State == "queued" {
		return ErrState
	}
	_, err = s.db.Exec(`DELETE FROM jobs WHERE id=? AND owner=?`, id, owner)
	return err
}
func (s *Store) Expired(now time.Time) ([]Job, error) {
	rows, err := s.db.Query(`SELECT body,url,option_id,owner FROM jobs WHERE expires<=?`, now.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Job{}
	for rows.Next() {
		j, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (s *Store) Recover() error {
	rows, err := s.db.Query(`SELECT body,url,option_id,owner FROM jobs WHERE state='processing'`)
	if err != nil {
		return err
	}
	list := []Job{}
	for rows.Next() {
		j, err := scan(rows)
		if err != nil {
			_ = rows.Close()
			return err
		}
		list = append(list, j)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for _, j := range list {
		_, err = s.Transition(j.ID, j.Owner, []string{"processing"}, func(j *Job) {
			j.State = "failed"
			j.Phase = "failed"
			j.Error = "Processing was interrupted by a restart. Retry to start again."
		})
		if err != nil {
			return err
		}
	}
	return nil
}
