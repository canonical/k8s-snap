package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyIngress is used to configure the ingress controller feature on Canonical Kubernetes.
// ApplyIngress assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyIngress will enable Cilium's ingress controller when ingress.Enabled is true.
// ApplyIngress will disable Cilium's ingress controller when ingress.Disabled is false.
// ApplyIngress will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyIngress returns an error if anything fails.
func ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network) error {
	m := snap.HelmClient()

	var values map[string]any
	if ingress.GetEnabled() {
		values = map[string]any{
			"ingressController": map[string]any{
				"enabled":                true,
				"loadbalancerMode":       "shared",
				"defaultSecretNamespace": "kube-system",
				"defaultTLSSecret":       ingress.GetDefaultTLSSecret(),
				"enableProxyProtocol":    ingress.GetEnableProxyProtocol(),
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

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values)
	if err != nil {
		return fmt.Errorf("failed to enable ingress: %w", err)
	}
	if !changed || !ingress.GetEnabled() {
		return nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply ingress: %w", err)
	}
	return nil
}
