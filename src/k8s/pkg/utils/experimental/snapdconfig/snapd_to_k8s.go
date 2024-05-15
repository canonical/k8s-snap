package snapdconfig

import (
	"context"
	"encoding/json"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/snap"
)

// SetK8sdFromSnapd updates the k8sd cluster configuration from the current local snapd configuration.
func SetK8sdFromSnapd(ctx context.Context, client client.Client, snap snap.Snap) error {
	b, err := snap.SnapctlGet(ctx, "-d", "dns", "network", "local-storage", "load-balancer", "ingress", "gateway")
	if err != nil {
		return fmt.Errorf("failed to retrieve snapd configuration: %w", err)
	}

	var config apiv1.UserFacingClusterConfig
	if err := json.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("failed to parse snapd configuration: %w", err)
	}

	if err := client.UpdateClusterConfig(ctx, apiv1.UpdateClusterConfigRequest{Config: config}); err != nil {
		return fmt.Errorf("failed to update k8s configuration: %w", err)
	}

	return nil
}
