package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

func main() {
	address := os.Getenv("OD_EGRESS_LISTEN")
	if address == "" {
		address = "127.0.0.1:8090"
	}
	server := &http.Server{Addr: address, Handler: security.NewProxy(security.Policy{}), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	slog.Info("guarded egress listening", "address", address)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("egress stopped", "error", err)
		os.Exit(1)
	}
}
