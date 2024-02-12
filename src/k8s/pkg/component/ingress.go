package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func EnableIngressComponent(s snap.Snap, defaultTLSSecret string, enableProxyProtocol bool, ctx context.Context) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	values := map[string]any{
		"ingressController": map[string]any{
			"enabled":                true,
			"loadbalancerMode":       "shared",
			"defaultSecretNamespace": "kube-system",
			"defaultTLSSecret":       defaultTLSSecret,
			"enableProxyProtocol":    enableProxyProtocol,
		},
	}

	err = manager.Refresh("network", values)
	if err != nil {
		return fmt.Errorf("failed to enable ingress component: %w", err)
	}

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

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

func DisableIngressComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	values := map[string]any{
		"ingressController": map[string]any{
			"enabled":                false,
			"defaultSecretNamespace": "",
			"defaultSecretName":      "",
			"enableProxyProtocol":    false,
		},
	}
	err = manager.Refresh("network", values)
	if err != nil {
		return fmt.Errorf("failed to disable ingress component: %w", err)
	}

	return nil
}
