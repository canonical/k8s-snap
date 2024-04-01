package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
)

func UpdateGatewayComponent(ctx context.Context, s snap.Snap, isRefresh bool) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	var values map[string]any = nil

	if isRefresh {
		if err := manager.Refresh("gateway", values); err != nil {
			return fmt.Errorf("failed to refresh gateway component: %w", err)
		}
	} else {
		if err := manager.Enable("gateway", values); err != nil {
			return fmt.Errorf("failed to enable gateway component: %w", err)
		}
	}

	networkValues := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": true,
		},
	}

	if err = manager.Refresh("network", networkValues); err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}

	client, err := k8s.NewClient(s.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// There is a race condition where the cilium resources can change
	// while we try to restart them, which fails with:
	// the object has been modified; please apply your changes to the latest version and try again
	if err := control.RetryFor(3, func() error {
		if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := control.RetryFor(3, func() error {
		if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func DisableGatewayComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	networkValues := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": false,
		},
	}

	if err := manager.Refresh("network", networkValues); err != nil {
		return fmt.Errorf("failed to disable gateway component: %w", err)
	}

	if err := manager.Disable("gateway"); err != nil {
		return fmt.Errorf("failed to disable gateway component: %w", err)
	}

	return nil
}

func ReconcileGatewayComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateGatewayComponent(ctx, s, true)
		if err != nil {
			return fmt.Errorf("failed to refresh gateway: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateGatewayComponent(ctx, s, false)
		if err != nil {
			return fmt.Errorf("failed to enable gateway: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableGatewayComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable gateway: %w", err)
		}
		return nil
	}
	return nil
}
