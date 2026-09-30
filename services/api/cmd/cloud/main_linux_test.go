//go:build linux

package main

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBridgeStreamsAndClosesOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	target := filepath.Join(t.TempDir(), "target.sock")
	upstream, err := net.Listen("unix", target)
	if err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	go func() {
		c, err := upstream.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = io.Copy(c, c)
	}()
	listener, err := bridge(ctx, "tcp", "127.0.0.1:0", "unix", target, cancel)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	c, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err = c.Write([]byte("media-stream")); err != nil {
		t.Fatal(err)
	}
	result := make([]byte, 12)
	if _, err = io.ReadFull(c, result); err != nil || string(result) != "media-stream" {
		t.Fatalf("echo=%q err=%v", result, err)
	}
	cancel()
	if _, err = c.Read(result); err == nil {
		t.Fatal("connection survived cancellation")
	}
	if _, err = listener.Accept(); err == nil {
		t.Fatal("listener survived cancellation")
	}
}

func TestListenPreservesNonSocketFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keep")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := listen("unix", path); err == nil {
		t.Fatal("replaced ordinary file")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "keep" {
		t.Fatalf("file=%q err=%v", data, err)
	}
}

func TestWorkerRefusesUnchangedNamespace(t *testing.T) {
	ns, err := namespace()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OD_CLOUD_HOST_NS", ns)
	if err := isolation(); err == nil {
		t.Fatal("worker accepted host network namespace")
	}
	t.Setenv("OD_CLOUD_HOST_NS", "")
	if err := isolation(); err == nil {
		t.Fatal("worker accepted missing namespace identity")
	}
}
