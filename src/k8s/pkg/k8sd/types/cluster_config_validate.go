package types

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
)

func validateCIDRs(cidrString string) error {
	cidrs := strings.Split(cidrString, ",")
	if v := len(cidrs); v != 1 && v != 2 {
		return fmt.Errorf("must contain 1 or 2 CIDRs, but found %d instead", v)
	}
	for _, cidr := range cidrs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("%q is not a valid CIDR: %w", cidr, err)
		}
	}
	return nil
}

// validateCIDROverlapAndSize checks for overlap and size constraints between pod and service CIDRs.
// It parses the provided podCIDR and serviceCIDR strings, checks for IPv4 and IPv6 overlaps.
func validateCIDROverlap(podCIDR string, serviceCIDR string) error {
	// Parse the CIDRs
	podIPv4CIDR, podIPv6CIDR, err := utils.SplitCIDRStrings(podCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse pod CIDR: %w", err)
	}

	svcIPv4CIDR, svcIPv6CIDR, err := utils.SplitCIDRStrings(serviceCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse service CIDR: %w", err)
	}

	// Check for IPv4 overlap
	if podIPv4CIDR != "" && svcIPv4CIDR != "" {
		if overlap, err := utils.CIDRsOverlap(podIPv4CIDR, svcIPv4CIDR); err != nil {
			return fmt.Errorf("failed to check for IPv4 overlap: %w", err)
		} else if overlap {
			return fmt.Errorf("pod CIDR %q and service CIDR %q overlap", podCIDR, serviceCIDR)
		}
	}

	// Check for IPv6 overlap
	if podIPv6CIDR != "" && svcIPv6CIDR != "" {
		if overlap, err := utils.CIDRsOverlap(podIPv6CIDR, svcIPv6CIDR); err != nil {
			return fmt.Errorf("failed to check for IPv6 overlap: %w", err)
		} else if overlap {
			return fmt.Errorf("pod CIDR %q and service CIDR %q overlap", podCIDR, serviceCIDR)
		}
	}

	return nil
}

// Check CIDR size ensures that the service IPv6 CIDR is not larger than /108.
// Ref: https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/networking/dualstack/#cidr-size-limitations
func validateIPv6CIDRSize(serviceCIDR string) error {
	_, svcIPv6CIDR, err := utils.SplitCIDRStrings(serviceCIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	if svcIPv6CIDR == "" {
		return nil
	}

	_, ipv6Net, err := net.ParseCIDR(svcIPv6CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	prefixLength, _ := ipv6Net.Mask.Size()
	if prefixLength < 108 {
		return fmt.Errorf("service CIDR %q cannot be larger than /108", serviceCIDR)
	}

	return nil
}

// Validate that a ClusterConfig does not have conflicting or incompatible options.
func (c *ClusterConfig) Validate() error {
	// check: validate that PodCIDR and ServiceCIDR are configured
	if err := validateCIDRs(c.Network.GetPodCIDR()); err != nil {
		return fmt.Errorf("invalid pod CIDR: %w", err)
	}
	if err := validateCIDRs(c.Network.GetServiceCIDR()); err != nil {
		return fmt.Errorf("invalid service CIDR: %w", err)
	}

	if err := validateCIDROverlap(c.Network.GetPodCIDR(), c.Network.GetServiceCIDR()); err != nil {
		return fmt.Errorf("invalid cidr configuration: %w", err)
	}
	// Can't be an else-if, because default values could already be set.
	if err := validateIPv6CIDRSize(c.Network.GetServiceCIDR()); err != nil {
		return fmt.Errorf("invalid service CIDR: %w", err)
	}

	// check: ensure network is enabled if any of ingress, gateway, load-balancer are enabled
	if !c.Network.GetEnabled() {
		if c.Gateway.GetEnabled() {
			return fmt.Errorf("gateway requires network to be enabled")
		}
		if c.LoadBalancer.GetEnabled() {
			return fmt.Errorf("load-balancer requires network to be enabled")
		}
		if c.Ingress.GetEnabled() {
			return fmt.Errorf("ingress requires network to be enabled")
		}
	}

	// check: load-balancer CIDRs
	for _, cidr := range c.LoadBalancer.GetCIDRs() {
		// Handle CIDR
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("load-balancer configuration contains an invalid CIDR %q: %w", cidr, err)
		}
	}

	for _, ipRange := range c.LoadBalancer.GetIPRanges() {
		start, err := netip.ParseAddr(ipRange.Start)
		if err != nil {
			return fmt.Errorf("load-balancer configuration contains an IP range (%#v) with an invalid start IP: %w", ipRange, err)
		}
		stop, err := netip.ParseAddr(ipRange.Stop)
		if err != nil {
			return fmt.Errorf("load-balancer configuration contains an IP range (%#v) with an invalid stop IP: %w", ipRange, err)
		}

		// Check if stop is greater than start
		if stop.Less(start) {
			return fmt.Errorf("load-balancer configuration contains an IP range (%#v) with start IP greater than the stop IP", ipRange)
		}
	}

	// check: load-balancer BGP mode configuration
	if c.LoadBalancer.GetBGPMode() {
		if c.LoadBalancer.GetBGPLocalASN() == 0 {
			return fmt.Errorf("load-balancer.bgp-local-asn must be set when load-balancer.bgp-mode is enabled")
		}
		if c.LoadBalancer.GetBGPPeerAddress() == "" {
			return fmt.Errorf("load-balancer.bgp-peer-address must be set when load-balancer.bgp-mode is enabled")
		}
		if c.LoadBalancer.GetBGPPeerPort() == 0 {
			return fmt.Errorf("load-balancer.bgp-peer-port must be set when load-balancer.bgp-mode is enabled")
		}
		if c.LoadBalancer.GetBGPPeerASN() == 0 {
			return fmt.Errorf("load-balancer.bgp-peer-asn must be set when load-balancer.bgp-mode is enabled")
		}
	}

	// check: local-storage.reclaim-policy should be one of 3 values
	switch c.LocalStorage.GetReclaimPolicy() {
	case "", "Retain", "Recycle", "Delete":
	default:
		return fmt.Errorf("local-storage.reclaim-policy must be one of: Retrain, Recycle, Delete")
	}

	// check: local-storage.local-path must be set if enabled
	if c.LocalStorage.GetEnabled() && c.LocalStorage.GetLocalPath() == "" {
		return fmt.Errorf("local-storage.local-path must be set when local-storage is enabled")
	}

	// check: ensure cluster DNS is a valid IP address
	if v := c.Kubelet.GetClusterDNS(); v != "" {
		if net.ParseIP(v) == nil {
			return fmt.Errorf("dns.service-ip must be a valid IP address")
		}

		// TODO: ensure dns.service-ip is part of new.Network.ServiceCIDR
	}

	// check: all external datastore servers are valid URLs
	for _, server := range c.Datastore.GetExternalServers() {
		if _, err := url.Parse(server); err != nil {
			return fmt.Errorf("datastore.external-servers contains invalid address: %s", server)
		}
	}

	return nil
}
