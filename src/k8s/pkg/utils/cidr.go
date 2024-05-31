package utils

import (
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	"github.com/canonical/lxd/lxd/util"
)

// findMatchingNodeAddress returns the IP address of a network interface that belongs to the given CIDR.
func findMatchingNodeAddress(cidr *net.IPNet) (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("could not get interface addresses: %w", err)
	}

	var selectedIP net.IP
	selectedSubnetBits := -1

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if cidr.Contains(ipNet.IP) {
			_, subnetBits := cidr.Mask.Size()
			if selectedSubnetBits == -1 || subnetBits < selectedSubnetBits {
				// Prefer the address with the fewest subnet bits
				selectedIP = ipNet.IP
				selectedSubnetBits = subnetBits
			}
		}
	}

	if selectedIP == nil {
		return nil, fmt.Errorf("could not find a matching address for CIDR %q", cidr.String())
	}

	return selectedIP, nil
}

// GetFirstIP returns the first IP address of a subnet. Use big.Int so that it can handle both IPv4 and IPv6 addreses.
func GetFirstIP(subnet string) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid subnet CIDR: %w", subnet, err)
	}
	r := big.NewInt(0).Add(
		big.NewInt(0).SetBytes(cidr.IP.To16()),
		big.NewInt(1),
	).Bytes()
	r = append(make([]byte, 16), r...)
	return net.IP(r[len(r)-16:]), nil
}

// GetKubernetesServiceIPsFromServiceCIDRs returns a list of the first IP addrs from a given service cidr string.
func GetKubernetesServiceIPsFromServiceCIDRs(serviceCIDR string) ([]net.IP, error) {
	var firstIPs []net.IP
	cidrs := strings.Split(serviceCIDR, ",")
	if v := len(cidrs); v != 1 && v != 2 {
		return nil, fmt.Errorf("invalid ServiceCIDR value: %v", cidrs)
	}
	for _, cidr := range cidrs {
		ip, err := GetFirstIP(cidr)
		if err != nil {
			return nil, fmt.Errorf("could not get IP from CIDR %q: %w", cidr, err)
		}
		firstIPs = append(firstIPs, ip)
	}
	return firstIPs, nil
}

// ParseAddressString parses an address string and returns a canonical network address.
func ParseAddressString(address string, port int64) (string, error) {
	host, hostPort, err := net.SplitHostPort(address)
	if err == nil {
		address = host
		port, err = strconv.ParseInt(hostPort, 10, 64)
		if err != nil {
			return "", fmt.Errorf("failed to parse the port from %q: %w", hostPort, err)
		}
	}

	if address == "" {
		address = util.NetworkInterfaceAddress()
	} else if _, ipNet, err := net.ParseCIDR(address); err == nil {
		matchingIP, err := findMatchingNodeAddress(ipNet)
		if err != nil {
			return "", fmt.Errorf("failed to find a matching node address for %q: %w", address, err)
		}
		address = matchingIP.String()
	}

	return util.CanonicalNetworkAddress(address, port), nil

}
