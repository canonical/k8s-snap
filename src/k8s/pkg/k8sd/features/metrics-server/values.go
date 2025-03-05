package metrics_server

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type Values map[string]any

func (v Values) applyDefaultValues() error {
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

func (v Values) ApplyImageOverrides(manifest types.FeatureManifest) error {
	metricsServerImage := manifest.GetImage(MetricsServerImageName)

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

	image := map[string]any{}

	values := map[string]any{
		"image": image,
	}

	if config.imageRepo != "" {
		image["repository"] = config.imageRepo
	}

	if config.imageTag != "" {
		image["tag"] = config.imageTag
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge annotation overrides: %w", err)
	}

	return nil
}
