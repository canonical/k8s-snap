package gateway

import (
	"fmt"

	"dario.cat/mergo"
)

type Values map[string]any

func (v Values) ApplyImageOverrides() error {
	ContourGatewayProvisionerContourImage := FeatureGateway.GetImage(ContourGatewayProvisionerContourImageName)
	ContourGatewayProvisionerEnvoyImage := FeatureGateway.GetImage(ContourGatewayProvisionerEnvoyImageName)

	values := map[string]any{
		"projectcontour": map[string]any{
			"image": map[string]any{
				"repository": ContourGatewayProvisionerContourImage.GetURI(),
				"tag":        ContourGatewayProvisionerContourImage.Tag,
			},
		},
		"envoyproxy": map[string]any{
			"image": map[string]any{
				"repository": ContourGatewayProvisionerEnvoyImage.GetURI(),
				"tag":        ContourGatewayProvisionerEnvoyImage.Tag,
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}
