package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when gateway.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when gateway.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway returns an error if anything fails.
func ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	var values map[string]any
	if gateway.GetEnabled() { //TODO: Do we need to chek for ingress enabled?
		values = map[string]any{
			"gateway": map[string]any{
				"gatewayRef": map[string]any{
					"name":      "gateway",
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
		return fmt.Errorf("failed to apply Gateway API contour configuration: %w", err)
	}

	if !changed || !gateway.GetEnabled() {
		return nil
	}
	if err := rolloutRestartContour(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart contour to apply Gateway API: %w", err)
	}

	return nil
}
