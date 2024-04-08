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

func UpdateIngressComponent(ctx context.Context, s snap.Snap, isRefresh bool, defaultTLSSecret string, enableProxyProtocol bool) error {
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

	client, err := k8s.NewClient(s.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	attempts := 3
	if err := control.RetryFor(attempts, func() error {
		if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(attempts, func() error {
		if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium daemonset after %d attempts: %w", attempts, err)
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

func ReconcileIngressComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateIngressComponent(ctx, s, true, clusterConfig.Ingress.GetDefaultTLSSecret(), clusterConfig.Ingress.GetEnableProxyProtocol())
		if err != nil {
			return fmt.Errorf("failed to refresh ingress: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateIngressComponent(ctx, s, false, clusterConfig.Ingress.GetDefaultTLSSecret(), clusterConfig.Ingress.GetEnableProxyProtocol())
		if err != nil {
			return fmt.Errorf("failed to enable ingress: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableIngressComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable ingress: %w", err)
		}
		return nil
	}
	return nil
}
