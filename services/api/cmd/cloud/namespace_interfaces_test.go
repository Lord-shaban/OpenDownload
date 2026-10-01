package main

import (
	"errors"
	"net"
	"testing"
)

func TestIsolatedInterfacesRejectNetworkAccessAndInspectionFailure(t *testing.T) {
	lo := net.Interface{Name: "lo", Flags: net.FlagUp | net.FlagLoopback | net.FlagRunning}
	fallbacks := []net.Interface{lo, {Name: "tunl0"}, {Name: "sit0"}, {Name: "ip6tnl0"}}
	for _, item := range []struct {
		name       string
		interfaces []net.Interface
		addrs      []net.Addr
		err        error
		allowed    bool
	}{
		{name: "loopback", interfaces: []net.Interface{lo}, allowed: true},
		{name: "inert kernel defaults", interfaces: fallbacks, allowed: true},
		{name: "unknown inactive interface", interfaces: []net.Interface{lo, {Name: "eth0"}}},
		{name: "active fallback", interfaces: []net.Interface{lo, {Name: "sit0", Flags: net.FlagUp}}},
		{name: "running fallback", interfaces: []net.Interface{lo, {Name: "tunl0", Flags: net.FlagRunning}}},
		{name: "addressed fallback", interfaces: fallbacks, addrs: []net.Addr{&net.IPNet{IP: net.ParseIP("192.0.2.1"), Mask: net.CIDRMask(24, 32)}}},
		{name: "inspection failure", interfaces: fallbacks, err: errors.New("netlink unavailable")},
		{name: "missing loopback"},
		{name: "inactive loopback", interfaces: []net.Interface{{Name: "lo", Flags: net.FlagLoopback}}},
		{name: "duplicate loopback", interfaces: []net.Interface{lo, lo}},
	} {
		t.Run(item.name, func(t *testing.T) {
			err := isolatedInterfaces(item.interfaces, func(net.Interface) ([]net.Addr, error) { return item.addrs, item.err })
			if (err == nil) != item.allowed {
				t.Fatalf("allowed=%v error=%v", item.allowed, err)
			}
		})
	}
}
