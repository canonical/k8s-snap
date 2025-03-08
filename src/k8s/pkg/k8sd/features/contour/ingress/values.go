package ingress

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
		"envoy-service-namespace": "projectcontour",
		"envoy-service-name":      "envoy",
		"contour": map[string]any{
			"manageCRDs": false,
			"ingressClass": map[string]any{
				"name":    "ck-ingress",
				"create":  true,
				"default": true,
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyImageOverrides() error {
	contourIngressContourImage := FeatureIngress.GetImage(ContourIngressContourImageName)
	contourIngressEnvoyImage := FeatureIngress.GetImage(ContourIngressEnvoyImageName)

	values := map[string]any{
		"envoy": map[string]any{
			"image": map[string]any{
				"registry":   "",
				"repository": contourIngressEnvoyImage.GetURI(),
				"tag":        contourIngressEnvoyImage.Tag,
			},
		},
		"contour": map[string]any{
			"image": map[string]any{
				"registry":   "",
				"repository": contourIngressContourImage.GetURI(),
				"tag":        contourIngressContourImage.Tag,
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(ingress types.Ingress) error {
	var values map[string]any
	if ingress.GetEnableProxyProtocol() {
		values = map[string]any{
			"contour": map[string]any{
				"extraArgs": []string{"--use-proxy-protocol"},
			},
		}
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}

type TLSValues map[string]any

func (v TLSValues) applyClusterConfiguration(ingress types.Ingress) error {
	values := map[string]any{
		"defaultTLSSecret": ingress.GetDefaultTLSSecret(),
	}

	if err := mergo.Merge(&v, TLSValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}
