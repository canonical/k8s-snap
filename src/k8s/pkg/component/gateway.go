package component

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func EnableGatewayComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	var values map[string]any = nil

	err = manager.Enable("gateway", values)
	if err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}

	networkValues := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": true,
		},
	}

	err = manager.Refresh("network", networkValues)
	if err != nil {
		return fmt.Errorf("failed to enable gateway component: %w", err)
	}

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = k8s.RestartDeployment(ctx, client, "cilium-operator", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
	}

	err = k8s.RestartDaemonset(ctx, client, "cilium", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
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

	err = manager.Refresh("network", networkValues)
	if err != nil {
		return fmt.Errorf("failed to disable gateway component: %w", err)
	}

	err = manager.Disable("gateway")
	if err != nil {
		return fmt.Errorf("failed to disable gateway component: %w", err)
	}

	return nil
}
