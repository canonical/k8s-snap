package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

func ApplyIngress(ctx context.Context, snap snap.Snap, cfg types.Ingress) error {
	m := newHelm(snap)

	var values map[string]any
	if cfg.GetEnabled() {
		values = map[string]any{
			"ingressController": map[string]any{
				"enabled":                true,
				"loadbalancerMode":       "shared",
				"defaultSecretNamespace": "kube-system",
				"defaultTLSSecret":       cfg.GetDefaultTLSSecret(),
				"enableProxyProtocol":    cfg.GetEnableProxyProtocol(),
			},
		}
	} else {
		values = map[string]any{
			"ingressController": map[string]any{
				"enabled":                false,
				"defaultSecretNamespace": "",
				"defaultSecretName":      "",
				"enableProxyProtocol":    false,
			},
		}
	}

	changed, err := m.Apply(ctx, featureNetwork, stateUpgradeOnly, values)
	if err != nil {
		return fmt.Errorf("failed to enable ingress: %w", err)
	}
	if !changed || !cfg.GetEnabled() {
		return nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply ingress: %w", err)
	}
	return nil
}
