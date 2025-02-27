package loadbalancer

import (
	"fmt"

	"dario.cat/mergo"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type CiliumValues cilium_network.Values

func (v CiliumValues) applyDefaultValues() error {
	values := map[string]any{
		"externalIPs": map[string]any{
			"enabled": true,
		},
		// https://docs.cilium.io/en/v1.14/network/l2-announcements/#sizing-client-rate-limit
		// Assuming for 50 LB services
		"k8sClientRateLimit": map[string]any{
			"qps":   10,
			"burst": 20,
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v CiliumValues) applyClusterConfiguration(loadbalancer types.LoadBalancer) error {
	values := map[string]any{
		"l2announcements": map[string]any{
			"enabled": loadbalancer.GetL2Mode(),
		},
		"bgpControlPlane": map[string]any{
			"enabled": loadbalancer.GetBGPMode(),
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v CiliumValues) applyDisableConfiguration() error {
	values := map[string]any{
		"l2announcements": map[string]any{
			"enabled": false,
		},
		"bgpControlPlane": map[string]any{
			"enabled": false,
		},
		"externalIPs": map[string]any{
			"enabled": false,
		},
		// https://docs.cilium.io/en/v1.14/network/l2-announcements/#sizing-client-rate-limit
		// Setting back to defaults
		"k8sClientRateLimit": map[string]any{
			"qps":   5,
			"burst": 10,
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
		"driver": "cilium",
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(loadbalancer types.LoadBalancer) error {
	cidrs := []map[string]any{}
	for _, cidr := range loadbalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range loadbalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := map[string]any{
		"l2": map[string]any{
			"enabled":    loadbalancer.GetL2Mode(),
			"interfaces": loadbalancer.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  loadbalancer.GetBGPMode(),
			"localASN": loadbalancer.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}
