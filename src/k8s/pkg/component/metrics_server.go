package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/vals"
)

func UpdateMetricsServerComponent(ctx context.Context, s snap.Snap, isRefresh bool) error {
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

	if isRefresh {
		if err := manager.Refresh("metrics-server", values); err != nil {
			return fmt.Errorf("failed to enable metrics-server component: %w", err)
		}
	} else {
		if err := manager.Enable("metrics-server", values); err != nil {
			return fmt.Errorf("failed to refresh metrics-server component: %w", err)
		}
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

func ReconcileMetricsServerComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateMetricsServerComponent(ctx, s, true)
		if err != nil {
			return fmt.Errorf("failed to refresh metrics-server: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateMetricsServerComponent(ctx, s, false)
		if err != nil {
			return fmt.Errorf("failed to enable metrics-server: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableMetricsServerComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable metrics-server: %w", err)
		}
		return nil
	}
	return nil
}
