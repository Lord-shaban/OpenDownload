package jobs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageReservationsIncludeRetainedFilesAndParallelJobs(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "retained"), make([]byte, 100), 0600); err != nil {
		t.Fatal(err)
	}
	m := &Manager{Root: root, MaxBytes: 100, StorageBudget: 300}
	first, err := m.reserveStorage()
	if err != nil {
		t.Fatal(err)
	}
	second, err := m.reserveStorage()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.reserveStorage(); err == nil {
		t.Fatal("overcommitted retained + reserved storage")
	}
	first()
	third, err := m.reserveStorage()
	if err != nil {
		t.Fatal("completed reservations must be released", err)
	}
	second()
	third()
	if err := os.Remove(filepath.Join(root, "retained")); err != nil {
		t.Fatal(err)
	}
	if m.reservedBytes != 0 {
		t.Fatal("reservation leaked")
	}
}
