package utils

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

// ParseEndpoints processes the given kube-apiserver endpoints and returns a list of
// IPv4:port or [IPv6]:port strings.
func ParseEndpoints(endpoints *corev1.Endpoints) []string {
	addresses := make([]string, 0, len(endpoints.Subsets))

	for _, subset := range endpoints.Subsets {
		portNumber := 6443
		for _, port := range subset.Ports {
			if port.Name == "https" {
				portNumber = int(port.Port)
				break
			}
		}

		for _, addr := range subset.Addresses {
			if addr.IP != "" {
				var address string
				if IsIPv4(addr.IP) {
					address = addr.IP
				} else {
					address = fmt.Sprintf("[%s]", addr.IP)
				}
				addresses = append(addresses, fmt.Sprintf("%s:%d", address, portNumber))
			}
		}
	}

	sort.Strings(addresses)
	return addresses
}
