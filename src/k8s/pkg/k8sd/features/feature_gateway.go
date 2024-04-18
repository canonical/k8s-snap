package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

func ApplyGateway(ctx context.Context, snap snap.Snap, cfg types.Gateway) error {
	m := newHelm(snap)

	if _, err := m.Apply(ctx, featureGateway, stateFromBool(cfg.GetEnabled()), nil); err != nil {
		return fmt.Errorf("failed to install Gateway API CRDs: %w", err)
	}

	changed, err := m.Apply(ctx, featureNetwork, stateUpgradeOnly, map[string]any{"gatewayAPI": map[string]any{"enabled": cfg.GetEnabled()}})
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
