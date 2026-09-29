package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}
func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	c, err := server.ConfigFromEnv()
	if err != nil {
		return err
	}
	data, err := filepath.Abs(c.Data)
	if err != nil {
		return err
	}
	store, err := jobs.Open(filepath.Join(data, "jobs.db"), c.Queue)
	if err != nil {
		return err
	}
	defer store.Close()
	var engine media.Engine
	if c.Fixture {
		engine = media.Fixture{Delay: 300 * time.Millisecond}
		slog.Warn("explicit fixture mode enabled; no real media is extracted")
	} else {
		engine, err = media.NewYTDLP(c.Binary, c.Proxy, c.MaxBytes)
		if err != nil {
			return err
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	manager := &jobs.Manager{Store: store, Engine: engine, Root: filepath.Join(data, "files"), Workers: c.Workers, MaxBytes: c.MaxBytes, Timeout: c.Timeout}
	if err := manager.Start(ctx); err != nil {
		return err
	}
	api := server.New(c, manager)
	httpServer := &http.Server{Addr: ":" + c.Port, Handler: api.Handler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	slog.Info("OpenDownload API listening", "port", c.Port, "workers", c.Workers)
	err = httpServer.ListenAndServe()
	stop()
	manager.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
