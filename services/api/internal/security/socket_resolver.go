package security

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

// SocketResolver supplies DNS validation to a network-isolated API. Only the
// guarded broker has access to the host resolver; the socket is never published.
type SocketResolver struct{ client *http.Client }

func NewSocketResolver(path string) *SocketResolver {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &SocketResolver{client: &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", path)
		}, MaxConnsPerHost: 8, MaxIdleConnsPerHost: 4,
	}}}
}

func (s *SocketResolver) LookupNetIP(ctx context.Context, _, host string) ([]netip.Addr, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://resolver/resolve?host="+url.QueryEscape(host), nil)
	if err != nil {
		return nil, ErrUnsafeURL
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrUnsafeURL
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnsafeURL
	}
	var ips []netip.Addr
	if json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&ips) != nil || len(ips) == 0 || len(ips) > 128 {
		return nil, ErrUnsafeURL
	}
	for _, ip := range ips {
		if !PublicIP(ip) {
			return nil, ErrUnsafeURL
		}
	}
	return ips, nil
}

// ResolverHandler never returns local/private DNS answers, including mixed sets.
func ResolverHandler(policy Policy) http.Handler {
	slots := make(chan struct{}, 8)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/resolve" {
			http.Error(w, "not found", 404)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "resolver busy", 503)
			return
		}
		host := r.URL.Query().Get("host")
		u, err := Parse("https://" + host)
		if err != nil || u.Hostname() != host || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			http.Error(w, "destination denied", 403)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		ips, err := policy.Resolve(ctx, host)
		if err != nil || len(ips) > 128 {
			http.Error(w, "destination denied", 403)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(ips)
	})
}
