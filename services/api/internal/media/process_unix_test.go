//go:build !windows

package media

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCancellationStopsDescendants(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := command(ctx, "sh", "-c", `(sleep 1; echo escaped > "$1/escaped") & echo ready > "$1/ready"; wait`, "sh", dir)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(dir, "ready")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := cmd.Wait(); err == nil {
		t.Fatal("canceled process succeeded")
	}
	time.Sleep(1200 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(dir, "escaped")); !os.IsNotExist(err) {
		t.Fatal("descendant survived cancellation")
	}
}
