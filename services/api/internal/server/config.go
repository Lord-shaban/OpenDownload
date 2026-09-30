package server

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	Data           string
	Origin         string
	Proxy          string
	Binary         string
	Workers        int
	Queue          int
	MaxBytes       int64
	Timeout        time.Duration
	Retention      time.Duration
	Fixture        bool
	ResolverSocket string
	StorageBudget  int64
}

func ConfigFromEnv() (Config, error) {
	c := Config{Port: value("OD_PORT", "8080"), Data: value("OD_DATA_DIR", ".data"), Origin: value("OD_ORIGIN", "http://localhost:3000"), Proxy: value("OD_EGRESS_PROXY", "http://127.0.0.1:8090"), Binary: value("OD_YTDLP", "yt-dlp")}
	var err error
	if c.Workers, err = number("OD_WORKERS", 2, 1, 8); err != nil {
		return c, err
	}
	if c.Queue, err = number("OD_QUEUE_LIMIT", 20, 1, 200); err != nil {
		return c, err
	}
	b, err := number("OD_MAX_BYTES", 536870912, 1024, 10<<30)
	if err != nil {
		return c, err
	}
	c.MaxBytes = int64(b)
	budget, err := number("OD_STORAGE_BUDGET", 0, 0, 10<<30)
	if err != nil || budget != 0 && int64(budget) < c.MaxBytes {
		return c, fmt.Errorf("OD_STORAGE_BUDGET must be zero or at least OD_MAX_BYTES, up to 10 GiB")
	}
	c.StorageBudget = int64(budget)
	c.ResolverSocket = os.Getenv("OD_RESOLVER_SOCKET")
	if c.Timeout, err = time.ParseDuration(value("OD_JOB_TIMEOUT", "15m")); err != nil || c.Timeout < time.Second || c.Timeout > time.Hour {
		return c, fmt.Errorf("OD_JOB_TIMEOUT must be between 1s and 1h")
	}
	if c.Retention, err = time.ParseDuration(value("OD_RETENTION", "1h")); err != nil || c.Retention < c.Timeout || c.Retention > 24*time.Hour {
		return c, fmt.Errorf("OD_RETENTION must be at least OD_JOB_TIMEOUT and at most 24h")
	}
	if c.Fixture, err = strconv.ParseBool(value("OD_FIXTURE_MODE", "false")); err != nil {
		return c, fmt.Errorf("OD_FIXTURE_MODE must be true or false")
	}
	u, err := url.Parse(c.Origin)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return c, fmt.Errorf("OD_ORIGIN must be an exact HTTP(S) origin")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return c, fmt.Errorf("OD_PORT is invalid")
	}
	return c, nil
}
func value(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
func number(name string, fallback, low, high int) (int, error) {
	v, err := strconv.Atoi(value(name, strconv.Itoa(fallback)))
	if err != nil || v < low || v > high {
		return 0, fmt.Errorf("%s must be between %d and %d", name, low, high)
	}
	return v, nil
}
