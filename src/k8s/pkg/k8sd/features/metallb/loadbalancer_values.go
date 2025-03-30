package metallb

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type metalLBValues map[string]any

func (v *metalLBValues) applyDefaults() error {
	values := metalLBValues{
		"controller": map[string]any{
			"command": "/controller",
		},
		"speaker": map[string]any{
			"command": "/speaker",
			// TODO(neoaggelos): make frr enable/disable configurable through an annotation
			// We keep it disabled by default
			"frr": map[string]any{
				"enabled": false,
			},
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *metalLBValues) applyImages() error {
	values := metalLBValues{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": controllerImageRepo,
				"tag":        ControllerImageTag,
			},
		},
		"speaker": map[string]any{
			"image": map[string]any{
				"repository": speakerImageRepo,
				"tag":        speakerImageTag,
			},
			"frr": map[string]any{
				"image": map[string]any{
					"repository": frrImageRepo,
					"tag":        frrImageTag,
				},
			},
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

type loadBalancerValues map[string]any

func (v *loadBalancerValues) applyDefaults() error {
	values := loadBalancerValues{
		"driver": "metallb",
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *loadBalancerValues) applyClusterConfig(loadbalancer types.LoadBalancer) error {
	cidrs := []map[string]any{}
	for _, cidr := range loadbalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range loadbalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := loadBalancerValues{
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

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
