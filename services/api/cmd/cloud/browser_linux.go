//go:build linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Only called by the private Unix diagnostic socket. No URL or flags come from
// the caller. Verify Chromium's implicit localhost bypass does not reach API.
func browserEgressCheck(ctx context.Context) error {
	const binary = "/usr/local/bin/opendownload-chromium"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return nil // Older/non-browser cloud images still have the Go fence check.
	}
	dir, err := os.MkdirTemp("", "od-browser-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--headless=new", "--user-data-dir="+dir, "--dump-dom", "http://127.0.0.1:8080/api/v1/status")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 3 * time.Second
	output, err := cmd.Output()
	if err != nil || !strings.Contains(string(output), "destination denied") || strings.Contains(string(output), "fixtureMode") {
		return errors.New("browser must route localhost requests through the guarded proxy")
	}
	return nil
}
