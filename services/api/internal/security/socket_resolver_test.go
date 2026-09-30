package security

import (
	"context"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestResolverBrokerRejectsMixedDNSAndMalformedHosts(t *testing.T) {
	for _, tc := range []struct {
		host   string
		dns    fakeDNS
		status int
	}{
		{"example.com", fakeDNS{netip.MustParseAddr("1.1.1.1")}, 200},
		{"example.com", fakeDNS{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}, 403},
		{"localhost", nil, 403}, {"127.0.0.1", nil, 403}, {"example.com/secret", nil, 403},
	} {
		r := httptest.NewRequest("GET", "/resolve?host="+tc.host, nil)
		w := httptest.NewRecorder()
		ResolverHandler(Policy{Resolver: tc.dns}).ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s: %d", tc.host, w.Code)
		}
	}
	resolver := NewSocketResolver("does-not-exist")
	if _, err := resolver.LookupNetIP(context.Background(), "ip", "example.com"); err == nil {
		t.Fatal("resolver must not fall back to native DNS")
	}
}
