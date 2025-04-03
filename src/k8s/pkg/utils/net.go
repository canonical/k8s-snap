package utils

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"syscall"

	"github.com/vishvananda/netlink"
)

// IsLocalPortOpen checks if the given local port is already open or not.
func IsLocalPortOpen(port string) (bool, error) {
	// Without an address, Listen will listen on all addresses.
	if l, err := net.Listen("tcp", fmt.Sprintf(":%s", port)); errors.Is(err, syscall.EADDRINUSE) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		l.Close()
		return true, nil
	}
}

// GetIPv46Addresses returns an IPv4 and IPv6 pair if possible, from the local network interface which has the
// given IP address. This looks through all the interface's IPs to find a global unicast IP of the opposite
// type (IPv4 / IPv6) from the input IP.
// GetIPv46Addresses will return an array containing only the given IP address if the pair could not be found.
// If the given IP address cannot be found locally, an error will be returned.
func GetIPv46Addresses(ip net.IP) ([]net.IP, error) {
	ifaceIPs, err := getIPsFromIfaceWithIP(ip)
	if err != nil {
		return nil, err
	}

	// Return a [IPv4, IPv6] pair if possible.
	isIPv4 := ip.To4() != nil
	for _, ifaceIP := range ifaceIPs {
		isIfaceIPv4 := ifaceIP.To4() != nil
		if isIPv4 == isIfaceIPv4 {
			// They are the same type. Skip.
			continue
		}

		if ifaceIP.IsGlobalUnicast() {
			return []net.IP{ip, ifaceIP}, nil
		}
	}

	// Couldn't find pair. Return the given IP.
	return []net.IP{ip}, nil
}

func getIPsFromIfaceWithIP(ip net.IP) ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get local interfaces: %w", err)
	}

	var lastError error
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			// Continue, we may still find the right interface.
			// Return lastError if we don't find it.
			lastError = fmt.Errorf("failed to get local interface addresses: %w", err)
			continue
		}

		ifaceIPs, err := parseIPAddresses(addrs)
		if err != nil {
			// Continue, we may still find the right interface.
			// Return lastError if we don't find it.
			lastError = err
			continue
		}

		// Check if the given IP is in the list.
		if slices.ContainsFunc(ifaceIPs, ip.Equal) {
			return ifaceIPs, nil
		}
	}

	if lastError != nil {
		return nil, lastError
	}

	return nil, fmt.Errorf("failed to find a local interface associated with the node IP address '%s'", ip.String())
}

func parseIPAddresses(addrs []net.Addr) ([]net.IP, error) {
	ips := []net.IP{}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil {
			return nil, fmt.Errorf("failed to parse node IP address '%s'", addr.String())
		}
		ips = append(ips, ip)
	}

	return ips, nil
}

func VxlanDevices() ([]netlink.Vxlan, error) {
	var vxlanDevices []netlink.Vxlan

	links, err := netlink.LinkList()
	if err != nil {
		return vxlanDevices, fmt.Errorf("failed to list network links: %w", err)
	}

	for _, link := range links {
		if vxlan, ok := link.(*netlink.Vxlan); ok {
			vxlanDevices = append(vxlanDevices, *vxlan)
		}
	}

	return vxlanDevices, nil
}

func RemoveLink(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to find the link %s: %w", name, err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to remove the link %s: %w", name, err)
	}

	return nil
}
