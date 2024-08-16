package snapdconfig

import (
	"context"
	"encoding/json"
	"fmt"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// SetSnapdFromK8sd uses snapctl to update the local snapd configuration with the new k8sd cluster configuration.
func SetSnapdFromK8sd(ctx context.Context, config apiv1.UserFacingClusterConfig, snap snap.Snap) error {
	var sets []string

	for key, cfg := range map[types.FeatureName]any{
		"meta":                Meta{Orb: "snapd", APIVersion: "1.30"},
		features.DNS:          config.DNS,
		features.Network:      config.Network,
		features.LocalStorage: config.LocalStorage,
		features.LoadBalancer: config.LoadBalancer,
		features.Ingress:      config.Ingress,
		features.Gateway:      config.Gateway,
	} {
		b, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal %s config: %w", key, err)
		}
		sets = append(sets, fmt.Sprintf("%s=%s", key, string(b)))
	}

	if err := snap.SnapctlSet(ctx, sets...); err != nil {
		return fmt.Errorf("failed to set snapd configuration: %w", err)
	}

	return nil
}
