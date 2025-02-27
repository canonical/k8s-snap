package gateway

import (
	"fmt"

	"dario.cat/mergo"
)

type Values map[string]any

func (v Values) ApplyImageOverrides() error {
	contourGatewayProvisionerContourImage := FeatureGateway.GetImage(ContourGatewayProvisionerContourImageName)
	contourGatewayProvisionerEnvoyImage := FeatureGateway.GetImage(ContourGatewayProvisionerEnvoyImageName)

	values := map[string]any{
		"projectcontour": map[string]any{
			"image": map[string]any{
				"repository": contourGatewayProvisionerContourImage.GetURI(),
				"tag":        contourGatewayProvisionerContourImage.Tag,
			},
		},
		"envoyproxy": map[string]any{
			"image": map[string]any{
				"repository": contourGatewayProvisionerEnvoyImage.GetURI(),
				"tag":        contourGatewayProvisionerEnvoyImage.Tag,
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}
