package metrics_server

import (
	"fmt"

	"dario.cat/mergo"
)

type Values map[string]any

func (v Values) ApplyDefaultValues() error {
	values := map[string]any{
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyImageOverrides() error {
	metricsServerImage := FeatureMetricsServer.GetImage(MetricsServerImageName)

	values := map[string]any{
		"image": map[string]any{
			"repository": metricsServerImage.GetURI(),
			"tag":        metricsServerImage.Tag,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v Values) ApplyAnnotations(annotations map[string]string) error {
	config := internalConfig(annotations)

	values := map[string]any{
		"image": map[string]any{},
	}

	if config.imageRepo != "" {
		values["image"].(map[string]any)["repository"] = config.imageRepo
	}

	if config.imageTag != "" {
		values["image"].(map[string]any)["tag"] = config.imageTag
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge annotation overrides: %w", err)
	}

	return nil
}
