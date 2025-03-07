package loadbalancer

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type MetalLBValues map[string]interface{}

func (v MetalLBValues) applyDefaultValues() error {
	values := map[string]any{
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

	if err := mergo.Merge(&v, MetalLBValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v MetalLBValues) ApplyImageOverrides() error {
	metalLBControllerImage := FeatureLoadBalancer.GetImage(MetalLBControllerImageName)
	metalLBSpeakerImage := FeatureLoadBalancer.GetImage(MetalLBSpeakerImageName)
	frrImage := FeatureLoadBalancer.GetImage(FRRImageName)

	values := map[string]any{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": metalLBControllerImage.GetURI(),
				"tag":        metalLBControllerImage.Tag,
			},
		},
		"speaker": map[string]any{
			"image": map[string]any{
				"repository": metalLBSpeakerImage.GetURI(),
				"tag":        metalLBSpeakerImage.Tag,
			},
			"frr": map[string]any{
				"image": map[string]any{
					"repository": frrImage.GetURI(),
					"tag":        frrImage.Tag,
				},
			},
		},
	}

	if err := mergo.Merge(&v, MetalLBValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
		"driver": "metallb",
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
		return fmt.Errorf("failed to merge cluster configuration values: %w", err)
	}

	return nil
}
