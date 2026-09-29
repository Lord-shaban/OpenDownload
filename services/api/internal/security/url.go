// Package security enforces public HTTP destinations at validation and dial time.
package security

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var ErrUnsafeURL = errors.New("use a public HTTP or HTTPS URL without credentials or custom ports")

type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

type Policy struct {
	Resolver Resolver
	dial     func(context.Context, string, string) (net.Conn, error)
}

var reserved = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func PublicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	// Conservatively restrict IPv6 to globally allocated unicast space.
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range reserved {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func Parse(raw string) (*url.URL, error) {
	if len(raw) > 4096 || strings.ContainsAny(raw, "\r\n\t\\") {
		return nil, ErrUnsafeURL
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Opaque != "" || u.User != nil || u.Hostname() == "" {
		return nil, ErrUnsafeURL
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrUnsafeURL
	}
	if u.Port() != "" && !((u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443")) {
		return nil, ErrUnsafeURL
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") && !strings.Contains(host, ":") {
		return nil, ErrUnsafeURL
	}
	u.Fragment = ""
	return u, nil
}

func (p Policy) Resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	host = strings.TrimSuffix(host, ".")
	if ip, err := netip.ParseAddr(host); err == nil {
		if !PublicIP(ip) {
			return nil, ErrUnsafeURL
		}
		return []netip.Addr{ip.Unmap()}, nil
	}
	resolver := p.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ips, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, ErrUnsafeURL
	}
	for _, ip := range ips {
		if !PublicIP(ip) {
			return nil, ErrUnsafeURL
		}
	}
	return ips, nil
}

func (p Policy) Validate(ctx context.Context, raw string) (*url.URL, error) {
	u, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	if _, err = p.Resolve(ctx, u.Hostname()); err != nil {
		return nil, err
	}
	return u, nil
}

func (p Policy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != "80" && port != "443" {
		return nil, ErrUnsafeURL
	}
	ips, err := p.Resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	dial := p.dial
	if dial == nil {
		dial = dialer.DialContext
	}
	// Dial the validated literal IP. Never ask DNS a second time.
	var last error
	for _, ip := range ips {
		conn, dialErr := dial(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		last = dialErr
	}
	return nil, last
}
