package security

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Proxy is deliberately unauthenticated and must never have a public port.
type Proxy struct {
	Policy    Policy
	Transport *http.Transport
	Slots     chan struct{}
}

func NewProxy(p Policy) *Proxy {
	return &Proxy{Policy: p, Slots: make(chan struct{}, 64), Transport: &http.Transport{
		DialContext: p.DialContext, TLSHandshakeTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second, MaxIdleConns: 32,
		IdleConnTimeout: 30 * time.Second, DisableCompression: true,
	}}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case p.Slots <- struct{}{}:
		defer func() { <-p.Slots }()
	default:
		http.Error(w, "outbound capacity reached", http.StatusServiceUnavailable)
		return
	}
	if r.Method == http.MethodConnect {
		p.tunnel(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPost {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	if _, err := Parse(r.URL.String()); err != nil {
		http.Error(w, "destination denied", http.StatusForbidden)
		return
	}
	out := r.Clone(r.Context())
	out.RequestURI = ""
	stripHop(out.Header)
	out.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	resp, err := p.Transport.RoundTrip(out)
	if err != nil {
		http.Error(w, "outbound request refused", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	stripHop(resp.Header)
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func stripHop(h http.Header) {
	for _, name := range strings.Split(h.Get("Connection"), ",") {
		h.Del(strings.TrimSpace(name))
	}
	for _, name := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE", "Trailer", "Transfer-Encoding", "Upgrade"} {
		h.Del(name)
	}
}

func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request) {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || port != "443" || host == "" {
		http.Error(w, "destination denied", http.StatusForbidden)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	upstream, err := p.Policy.DialContext(ctx, "tcp", r.Host)
	cancel()
	if err != nil {
		http.Error(w, "destination denied", http.StatusForbidden)
		return
	}
	defer upstream.Close()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "tunnel unavailable", http.StatusInternalServerError)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer client.Close()
	deadline := time.Now().Add(16 * time.Minute)
	_ = client.SetDeadline(deadline)
	_ = upstream.SetDeadline(deadline)
	_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	if buffered.Flush() != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(upstream, buffered)
		_ = upstream.Close()
	}()
	_, _ = io.Copy(client, upstream)
	_ = client.Close()
	wg.Wait()
}
