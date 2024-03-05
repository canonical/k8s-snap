package utils

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

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
			return nil, fmt.Errorf("could not get IP from CIDR: %v", cidr)
		}
		firstIPs = append(firstIPs, ip)
	}
	return firstIPs, nil
}
