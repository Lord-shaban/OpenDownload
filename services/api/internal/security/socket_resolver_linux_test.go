//go:build linux

package security

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"testing"
)

func TestSocketResolverRoundtripAndFailClosed(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		valid      bool
	}{
		{"public", `["8.8.8.8"]`, true},
		{"mixed", `["8.8.8.8","127.0.0.1"]`, false},
		{"malformed", `["not-an-ip"]`, false},
		{"empty", `[]`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "dns.sock")
			l, err := net.Listen("unix", path)
			if err != nil {
				t.Fatal(err)
			}
			server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("host") != "example.org" {
					t.Error("hostname not carried to broker")
				}
				_, _ = w.Write([]byte(tt.body))
			})}
			go func() { _ = server.Serve(l) }()
			defer server.Close()
			resolver := NewSocketResolver(path)
			defer resolver.client.CloseIdleConnections()
			ips, err := resolver.LookupNetIP(context.Background(), "ip", "example.org")
			if tt.valid {
				if err != nil || len(ips) != 1 || ips[0] != netip.MustParseAddr("8.8.8.8") {
					t.Fatalf("ips=%v err=%v", ips, err)
				}
			} else if err == nil {
				t.Fatal("unsafe broker answer accepted")
			}
		})
	}
}
