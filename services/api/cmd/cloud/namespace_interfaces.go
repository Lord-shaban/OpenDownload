package main

import (
	"errors"
	"net"
)

// Linux can create these fallback tunnel devices in a fresh network namespace
// when their kernel modules are loaded. Accept only inert defaults; never an
// active interface, address, unknown device or failure to inspect the namespace.
// The caller still requires a changed network namespace, PID 1 and denied egress.
func isolatedInterfaces(interfaces []net.Interface, addresses func(net.Interface) ([]net.Addr, error)) error {
	loopbacks := 0
	for _, iface := range interfaces {
		if iface.Name == "lo" && iface.Flags&net.FlagLoopback != 0 && iface.Flags&net.FlagUp != 0 {
			loopbacks++
			continue
		}
		if iface.Name != "tunl0" && iface.Name != "sit0" && iface.Name != "ip6tnl0" {
			return errors.New("worker has an unexpected network interface")
		}
		if iface.Flags != 0 {
			return errors.New("worker fallback tunnel must remain inactive")
		}
		addrs, err := addresses(iface)
		if err != nil || len(addrs) != 0 {
			return errors.New("worker fallback tunnel must have no addresses")
		}
	}
	if loopbacks != 1 {
		return errors.New("worker must have exactly one active loopback interface")
	}
	return nil
}
