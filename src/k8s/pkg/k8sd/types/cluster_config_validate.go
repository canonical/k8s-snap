package types

import (
	"fmt"
	"net"
	"strings"
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

// Validate that a ClusterConfig does not have conflicting or incompatible options.
func (c *ClusterConfig) Validate() error {
	// check: validate that PodCIDR and ServiceCIDR are configured
	if err := validateCIDRs(c.Network.GetPodCIDR()); err != nil {
		return fmt.Errorf("invalid pod CIDR: %w", err)
	}
	if err := validateCIDRs(c.Network.GetServiceCIDR()); err != nil {
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

	return nil
}
