package metrics_server

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type metricsServerValues map[string]any

func (v *metricsServerValues) applyDefaults() error {
	values := metricsServerValues{
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *metricsServerValues) applyImages() error {
	values := metricsServerValues{
		"image": map[string]any{
			"repository": imageRepo,
			"tag":        imageTag,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *metricsServerValues) applyAnnotations(annotation types.Annotations) error {
	config := internalConfig(annotation)

	values := metricsServerValues{
		"image": map[string]any{
			"repository": config.imageRepo,
			"tag":        config.imageTag,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
