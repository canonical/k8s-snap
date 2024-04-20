package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyGateway is used to configure the gateway feature on Canonical Kubernetes.
// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when cfg.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when cfg.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway returns an error if anything fails.
func ApplyGateway(ctx context.Context, snap snap.Snap, cfg types.Gateway) error {
	m := newHelm(snap)

	if _, err := m.Apply(ctx, featureCiliumGateway, stateFromBool(cfg.GetEnabled()), nil); err != nil {
		return fmt.Errorf("failed to install Gateway API CRDs: %w", err)
	}

	changed, err := m.Apply(ctx, featureCiliumCNI, stateUpgradeOnly, map[string]any{"gatewayAPI": map[string]any{"enabled": cfg.GetEnabled()}})
	if err != nil {
		return fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
	}

	if !changed || !cfg.GetEnabled() {
		return nil
	}
	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply Gateway API: %w", err)
	}
	return nil
}
