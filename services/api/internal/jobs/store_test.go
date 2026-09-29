package jobs

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func testStore(t *testing.T, limit int) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "jobs.db"), limit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
func sample(owner string) Job {
	return Job{Owner: owner, Title: "Sample", URL: "https://example.com", OptionID: "video", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
}
func TestQueueOwnershipAndTerminalStates(t *testing.T) {
	s := testStore(t, 1)
	j, err := s.Create(sample("owner"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Create(sample("other")); err != ErrCapacity {
		t.Fatal("queue capacity not enforced")
	}
	if _, err = s.Get(j.ID, "other"); err != ErrNotFound {
		t.Fatal("owner leak")
	}
	claimed, err := s.Claim()
	if err != nil || claimed.ID != j.ID {
		t.Fatal("claim failed")
	}
	_, err = s.Transition(j.ID, "owner", []string{"processing"}, func(j *Job) { j.State = "canceled" })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Transition(j.ID, "owner", []string{"processing"}, func(j *Job) { j.State = "complete" }); err != ErrState {
		t.Fatal("canceled job completed")
	}
}
func TestClaimsAreExclusive(t *testing.T) {
	s := testStore(t, 20)
	for i := 0; i < 5; i++ {
		if _, err := s.Create(sample("owner")); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	ids := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if j, err := s.Claim(); err == nil {
				ids <- j.ID
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatal("job claimed twice")
		}
		seen[id] = true
	}
	if len(seen) != 5 {
		t.Fatal("missing jobs")
	}
}
func TestRestartRecovery(t *testing.T) {
	s := testStore(t, 2)
	j, _ := s.Create(sample("owner"))
	_, _ = s.Claim()
	if err := s.Recover(); err != nil {
		t.Fatal(err)
	}
	recovered, _ := s.Get(j.ID, "owner")
	if recovered.State != "failed" || recovered.Error == "" {
		t.Fatal("interrupted job not explicit")
	}
}
