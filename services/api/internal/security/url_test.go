package security

import (
	"context"
	"net/netip"
	"testing"
)

type fakeDNS []netip.Addr

func (f fakeDNS) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) { return f, nil }

func TestPublicIP(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.1.2.3", "169.254.169.254", "100.100.100.200", "0.0.0.0", "192.0.2.8", "224.0.0.1", "::1", "::ffff:127.0.0.1", "fe80::1", "fc00::1", "64:ff9b::a00:1", "2001:db8::1", "2002:7f00:1::1"} {
		if PublicIP(netip.MustParseAddr(raw)) {
			t.Errorf("accepted reserved IP %s", raw)
		}
	}
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "::ffff:8.8.8.8"} {
		if !PublicIP(netip.MustParseAddr(raw)) {
			t.Errorf("rejected public IP %s", raw)
		}
	}
}

func TestURLPolicy(t *testing.T) {
	p := Policy{Resolver: fakeDNS{netip.MustParseAddr("1.1.1.1")}}
	for _, raw := range []string{"file:///etc/passwd", "ftp://example.com/a", "https://user:pass@example.com", "https://localhost/a", "https://example.com:8080/a", "http://127.0.0.1", "https://[::1]/a", "https://example.com\\@localhost/a", "https://example.com/\nfoo"} {
		if _, err := p.Validate(context.Background(), raw); err == nil {
			t.Errorf("accepted unsafe URL %s", raw)
		}
	}
	if _, err := p.Validate(context.Background(), "https://example.com/a#fragment"); err != nil {
		t.Fatal(err)
	}
	p.Resolver = fakeDNS{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}
	if _, err := p.Validate(context.Background(), "https://example.com"); err == nil {
		t.Fatal("mixed DNS must fail closed")
	}
	if _, err := p.DialContext(context.Background(), "tcp", "example.com:22"); err == nil {
		t.Fatal("custom port allowed")
	}
}
