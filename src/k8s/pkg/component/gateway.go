package component

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/snap"
)

func EnableGatewayComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	var values map[string]any = nil

	if err := manager.Enable("gateway", values); err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}

	networkValues := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": true,
		},
	}

	if err = manager.Refresh("network", networkValues); err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}

	client, err := s.KubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
	}

	if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart cilium daemonset: %w", err)
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
