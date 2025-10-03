package utils

import (
	"fmt"
	"sort"

	discoveryv1 "k8s.io/api/discovery/v1"
)

// ParseEndpoints processes the given kube-apiserver endpointslices and returns a list of
// IPv4:port or [IPv6]:port strings.
func ParseEndpointSlices(endpointSlices *discoveryv1.EndpointSliceList) []string {
	addresses := make([]string, 0, len(endpointSlices.Items))

	for _, endpointSlice := range endpointSlices.Items {
		portNumber := 6443
		for _, port := range endpointSlice.Ports {
			if port.Name != nil && *port.Name == "https" {
				if port.Port != nil {
					portNumber = int(*port.Port)
					break
				}
			}
		}

		for _, endpoint := range endpointSlice.Endpoints {
			for _, addr := range endpoint.Addresses {
				if addr != "" {
					var address string
					switch endpointSlice.AddressType {
					case discoveryv1.AddressTypeIPv4:
						address = addr
					case discoveryv1.AddressTypeIPv6:
						address = fmt.Sprintf("[%s]", addr)
					case discoveryv1.AddressTypeFQDN:
						// Not supported, skip.
						continue
					default:
						// Unknown address type, skip.
						continue
					}
					addresses = append(addresses, fmt.Sprintf("%s:%d", address, portNumber))
				}
			}
		}

	}

	sort.Strings(addresses)
	return addresses
}
