package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

func EnableMetricsServerComponent(ctx context.Context, s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": metricsServerImageRepository,
			"tag":        metricsServerImageTag,
		},
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	if err := manager.Enable("metrics-server", values); err != nil {
		return fmt.Errorf("failed to enable metrics-server component: %w", err)
	}

	return nil
}

func DisableMetricsServerComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("metrics-server"); err != nil {
		return fmt.Errorf("failed to disable metrics-server component: %w", err)
	}

	return nil
}
