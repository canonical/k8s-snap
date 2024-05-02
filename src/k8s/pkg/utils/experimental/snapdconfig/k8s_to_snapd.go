package snapdconfig

import (
	"context"
	"encoding/json"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/snap"
)

func SetSnapdFromK8sd(ctx context.Context, client client.Client, snap snap.Snap) error {
	config, err := client.GetClusterConfig(ctx, apiv1.GetClusterConfigRequest{})
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	var sets []string
	for key, cfg := range map[string]any{
		"meta":          Meta{Orb: "k8sd", APIVersion: "1.30"},
		"dns":           config.DNS,
		"network":       config.Network,
		"local-storage": config.LocalStorage,
		"load-balancer": config.LoadBalancer,
		"ingress":       config.Ingress,
		"gateway":       config.Gateway,
	} {
		b, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal %s config: %w", err)
		}
		sets = append(sets, fmt.Sprintf("%s=%s", key, string(b)))
	}

	if err := snap.SnapctlSet(ctx, sets...); err != nil {
		return fmt.Errorf("failed to set snapd configuration: %w", err)
	}

	return nil
}
