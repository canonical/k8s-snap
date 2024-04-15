package types

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// loadBalancerCIDRsFromAPI splits a list of CIDRs from API and returns the parsed list of CIDRs and IP ranges.
// loadBalancerCIDRsFromAPI returns nil lists for nil input, but empty lists for a valid empty list.
// loadBalancerCIDRsFromAPI returns an error if the input is not valid.
func loadBalancerCIDRsFromAPI(inCIDRs *[]string) (*[]string, *[]LoadBalancer_IPRange, error) {
	if inCIDRs == nil {
		return nil, nil, nil
	}
	outCIDRs := []string{}
	outRanges := []LoadBalancer_IPRange{}
	if len(*inCIDRs) == 0 {
		return &outCIDRs, &outRanges, nil
	}

	for _, cidr := range *inCIDRs {
		// Handle IP Range
		if strings.Contains(cidr, "-") {
			ipRange := strings.Split(cidr, "-")
			if len(ipRange) != 2 {
				return nil, nil, fmt.Errorf("load-balancer.cidrs contains an IP range (%q) not in $START-$END format", cidr)
			}

			start, err := netip.ParseAddr(ipRange[0])
			if err != nil {
				return nil, nil, fmt.Errorf("load-balancer.cidrs contains an IP range (%q) with an invalid start IP (%q): %w", cidr, ipRange[0], err)
			}
			stop, err := netip.ParseAddr(ipRange[1])
			if err != nil {
				return nil, nil, fmt.Errorf("load-balancer.cidrs contains an IP range (%q) with an invalid stop IP (%q): %w", cidr, ipRange[1], err)
			}
			if stop.Less(start) {
				return nil, nil, fmt.Errorf("load-balancer.cidrs contains an IP range (%q) with start IP (%q) greater than the stop IP (%q)", cidr, ipRange[0], ipRange[1])
			}

			outRanges = append(outRanges, LoadBalancer_IPRange{Start: ipRange[0], Stop: ipRange[1]})
		} else {
			// Handle CIDR
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return nil, nil, fmt.Errorf("load-balancer.cidrs contains an invalid CIDR %q: %w", cidr, err)
			}

			outCIDRs = append(outCIDRs, cidr)
		}
	}

	return &outCIDRs, &outRanges, nil
}

// loadBalancerCIDRsToAPI encodes internal CIDR and IP ranges for the LoadBalancer.CIDRs API field.
// loadBalancerCIDRsToAPI returns a nil list of CIDRs if both inputs are nil, otherwise an empty list.
func loadBalancerCIDRsToAPI(inCIDRs *[]string, inRanges *[]LoadBalancer_IPRange) *[]string {
	if inCIDRs == nil && inRanges == nil {
		return nil
	}

	outCIDRs := []string{}

	if inCIDRs != nil {
		outCIDRs = append(outCIDRs, *inCIDRs...)
	}

	if inRanges != nil {
		for _, ipRange := range *inRanges {
			outCIDRs = append(outCIDRs, fmt.Sprintf("%s-%s", ipRange.Start, ipRange.Stop))
		}
	}

	return &outCIDRs
}
