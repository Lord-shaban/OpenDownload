//go:build linux

// Cloud preserves the egress fence inside one container using an unprivileged
// user/network namespace. Unix sockets carry ingress, proxy traffic and DNS
// validation across that boundary; no public routes exist in the worker.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

const socketDir = "/tmp/opendownload-cloud"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	var err error
	if len(os.Args) == 2 && os.Args[1] == "worker" {
		err = worker(ctx)
	} else if len(os.Args) == 1 {
		err = host(ctx)
	} else {
		err = errors.New("unexpected cloud command")
	}
	if err != nil {
		slog.Error("cloud startup or runtime failed", "error", err)
		os.Exit(1)
	}
}

func namespace() (string, error) { return os.Readlink("/proc/self/ns/net") }

func listen(network, address string) (net.Listener, error) {
	if network == "unix" {
		if info, err := os.Lstat(address); err == nil {
			if info.Mode()&os.ModeSocket == 0 {
				return nil, errors.New("refusing to replace a non-socket path")
			}
			if err := os.Remove(address); err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	l, err := net.Listen(network, address)
	if err == nil && network == "unix" {
		if err = os.Chmod(address, 0600); err != nil {
			_ = l.Close()
			return nil, err
		}
	}
	return l, err
}

// Bound connection count, lifetime and shutdown for both TCP/Unix bridges.
func bridge(ctx context.Context, network, address, targetNetwork, target string, stop context.CancelFunc) (net.Listener, error) {
	l, err := listen(network, address)
	if err != nil {
		return nil, err
	}
	go func() { <-ctx.Done(); _ = l.Close() }()
	slots := make(chan struct{}, 128)
	go func() {
		for {
			client, err := l.Accept()
			if err != nil {
				if ctx.Err() == nil {
					slog.Error("cloud bridge stopped", "address", address, "error", err)
					stop()
				}
				return
			}
			select {
			case slots <- struct{}{}:
			default:
				_ = client.Close()
				continue
			}
			go func() {
				defer func() { <-slots; _ = client.Close() }()
				upstream, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, targetNetwork, target)
				if err != nil {
					return
				}
				defer upstream.Close()
				deadline := time.Now().Add(16 * time.Minute)
				_ = client.SetDeadline(deadline)
				_ = upstream.SetDeadline(deadline)
				done := make(chan struct{})
				defer close(done)
				go func() {
					select {
					case <-ctx.Done():
						_ = client.Close()
						_ = upstream.Close()
					case <-done:
					}
				}()
				copied := make(chan struct{}, 1)
				go func() { _, _ = io.Copy(upstream, client); copied <- struct{}{} }()
				go func() { _, _ = io.Copy(client, upstream); copied <- struct{}{} }()
				<-copied
				_ = client.Close()
				_ = upstream.Close()
				<-copied
			}()
		}
	}()
	return l, nil
}

func serveSocket(ctx context.Context, name string, handler http.Handler, stop context.CancelFunc) error {
	l, err := listen("unix", filepath.Join(socketDir, name))
	if err != nil {
		return err
	}
	s := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	go func() { <-ctx.Done(); _ = s.Close() }()
	go func() {
		if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) && ctx.Err() == nil {
			slog.Error("cloud socket stopped", "socket", name, "error", err)
			stop()
		}
	}()
	return nil
}

func managed(ctx context.Context, cmd *exec.Cmd, ready func(<-chan struct{}) error) error {
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	var result error
	go func() { result = cmd.Wait(); close(done) }()
	defer func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(15 * time.Second):
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-done
		}
	}()
	if ready != nil {
		if err := ready(done); err != nil {
			return err
		}
	}
	select {
	case <-ctx.Done():
		return nil
	case <-done:
		if result == nil {
			return errors.New("cloud child stopped unexpectedly")
		}
		return result
	}
}

func host(ctx context.Context) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	if os.Geteuid() == 0 {
		return errors.New("cloud host must run as a non-root user")
	}
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return err
	}
	if err := serveSocket(ctx, "egress.sock", security.NewProxy(security.Policy{}), stop); err != nil {
		return err
	}
	if err := serveSocket(ctx, "resolver.sock", security.ResolverHandler(security.Policy{}), stop); err != nil {
		return err
	}
	before, err := namespace()
	if err != nil {
		return err
	}
	// No fallback to a shared network, host networking, or privileged mode.
	// Keep Docker's existing /proc mount: its masked paths make a nested proc
	// mount unavailable on ordinary engines. PID isolation still reaps children.
	cmd := exec.Command("unshare", "--user", "--map-root-user", "--net", "--pid", "--fork", "--kill-child=SIGTERM", "--", "sh", "-c", "ip link set lo up && exec /usr/local/bin/opendownload-cloud worker")
	cmd.Env = append(os.Environ(), "OD_CLOUD_HOST_NS="+before)
	return managed(ctx, cmd, func(done <-chan struct{}) error {
		client := &http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", filepath.Join(socketDir, "web.sock"))
		}}}
		defer client.CloseIdleConnections()
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-done:
				return errors.New("isolated worker stopped before readiness")
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			resp, err := client.Get("http://web/api/v1/status")
			if err == nil {
				var status struct {
					Ready, FixtureMode bool
					Dependencies       map[string]bool
				}
				decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 16384)).Decode(&status)
				_ = resp.Body.Close()
				if resp.StatusCode == 200 && decodeErr == nil && status.Ready && !status.FixtureMode && status.Dependencies["ytDlp"] && status.Dependencies["ffmpeg"] {
					_, err = bridge(ctx, "tcp", "0.0.0.0:3000", "unix", filepath.Join(socketDir, "web.sock"), stop)
					if err == nil {
						slog.Info("real cloud downloads ready", "port", 3000)
					}
					return err
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(250 * time.Millisecond):
			}
		}
		return errors.New("real cloud dependencies did not become ready")
	})
}

func isolation() error {
	current, err := namespace()
	if err != nil || current == os.Getenv("OD_CLOUD_HOST_NS") || os.Getenv("OD_CLOUD_HOST_NS") == "" {
		return errors.New("worker network namespace was not changed")
	}
	if os.Getpid() != 1 {
		return errors.New("worker PID namespace was not changed")
	}
	interfaces, err := net.Interfaces()
	if err != nil || len(interfaces) != 1 || interfaces[0].Flags&net.FlagLoopback == 0 {
		return errors.New("worker must have only the loopback interface")
	}
	conn, err := net.DialTimeout("tcp", "1.1.1.1:443", time.Second)
	if err == nil {
		_ = conn.Close()
		return errors.New("worker has direct internet egress")
	}
	return nil
}

func worker(ctx context.Context) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	if err := isolation(); err != nil {
		return err
	}
	slog.Info("cloud worker network isolated; direct egress denied")
	// The diagnostic socket is private and is never routed by the web server.
	if err := serveSocket(ctx, "check.sock", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := isolation(); err != nil {
			http.Error(w, "isolation failed", 500)
			return
		}
		proxy, _ := url.Parse("http://127.0.0.1:8090")
		client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
		defer client.CloseIdleConnections()
		for _, target := range []string{"http://169.254.169.254/latest/meta-data/", "https://127.0.0.1/"} {
			resp, err := client.Get(target)
			if err != nil {
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode != 403 && resp.StatusCode != 502 {
				http.Error(w, "private proxy destination was not refused", 500)
				return
			}
		}
		if err := browserEgressCheck(r.Context()); err != nil {
			http.Error(w, "browser egress check failed", 500)
			return
		}
		_, _ = fmt.Fprintln(w, "direct egress denied")
	}), stop); err != nil {
		return err
	}
	if _, err := bridge(ctx, "tcp", "127.0.0.1:8090", "unix", filepath.Join(socketDir, "egress.sock"), stop); err != nil {
		return err
	}
	if _, err := bridge(ctx, "unix", filepath.Join(socketDir, "web.sock"), "tcp", "127.0.0.1:3001", stop); err != nil {
		return err
	}
	return managed(ctx, exec.Command("node", "deploy/cloud-server.mjs"), nil)
}
