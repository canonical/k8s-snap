package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyGateway assumes that the Contour is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the envoy-gateway-system on the cluster and enable set the right gateway configs in contour when gateway.Enabled is true.
// ApplyGateway will remove the envoy-gateway-system from the cluster and remove the right gateway configs in contour when gateway.Enabled is false.
// ApplyGateway will rollout restart the Contour pods.
// ApplyGateway returns an error if anything fails.
func ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()
	// First Install envoy-gateway-system
	if gateway.GetEnabled() {
		if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, nil); err != nil {
			return fmt.Errorf("failed to install envoy-gateway-system: %w", err)
		}

	} else {
		if _, err := m.Apply(ctx, chartGateway, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall envoy-gateway-system: %w", err)
		}
	}

	// Second update gateway config bits in contour ingress
	var values map[string]any
	if gateway.GetEnabled() { //TODO: Do we need to check for ingress enabled? Are we overwriting values set in ingress?
		values = map[string]any{
			"gateway": map[string]any{
				"gatewayRef": map[string]any{
					"name":      "contour",
					"namespace": "projectcontour",
				},
			},
		}
	} else {
		values = map[string]any{
			"gateway": map[string]any{
				"gatewayRef": map[string]any{
					"name":      "",
					"namespace": "",
				},
			},
		}
	}
	changed, err := m.Apply(ctx, chartContour, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values)
	if err != nil {
		return fmt.Errorf("failed to apply Gateway configuration to contour: %w", err)
	}

	if !changed || !gateway.GetEnabled() {
		return nil
	}
	if err := rolloutRestartContour(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart contour to apply Gateway configuration: %w", err)
	}

	return nil
}
