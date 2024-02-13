package component

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func EnableIngressComponent(s snap.Snap, defaultTLSSecret string, enableProxyProtocol bool) error {
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

	if err := manager.Refresh("network", values); err != nil {
		return fmt.Errorf("failed to enable ingress component: %w", err)
	}

	client, err := k8s.NewClient(s)
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
	if err := manager.Refresh("network", values); err != nil {
		return fmt.Errorf("failed to disable ingress component: %w", err)
	}

	return nil
}
