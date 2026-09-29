package security

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"testing"
)

type switchingDNS struct{ calls int }

func (d *switchingDNS) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	d.calls++
	if d.calls > 1 {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
	}
	return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
}
func TestDialPinsValidatedLiteral(t *testing.T) {
	dns := &switchingDNS{}
	p := Policy{Resolver: dns, dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != "1.1.1.1:443" {
			t.Fatalf("unvalidated dial: %s %s", network, address)
		}
		return nil, errors.New("test dial")
	}}
	_, _ = p.DialContext(context.Background(), "tcp", "rebinding.example:443")
	if dns.calls != 1 {
		t.Fatalf("resolved %d times", dns.calls)
	}
}
func TestProxyRejectsRedirectToPrivateAndConnect(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer upstream.Close()
	upURL, _ := url.Parse(upstream.URL)
	p := Policy{Resolver: fakeDNS{netip.MustParseAddr("1.1.1.1")}, dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "1.1.1.1:80" {
			t.Fatalf("unexpected destination: %s", address)
		}
		return (&net.Dialer{}).DialContext(ctx, network, upURL.Host)
	}}
	proxy := NewProxy(p)
	defer proxy.Transport.CloseIdleConnections()
	server := httptest.NewServer(proxy)
	defer server.Close()
	proxyURL, _ := url.Parse(server.URL)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport}
	resp, err := client.Get("http://public.example/media")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("private redirect was not denied: %d", resp.StatusCode)
	}
	for _, target := range []string{"127.0.0.1:443", "169.254.169.254:443", "public.example:22"} {
		r := httptest.NewRequest(http.MethodConnect, "http://"+target, nil)
		r.Host = target
		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("CONNECT %s returned %d", target, w.Code)
		}
	}
}
