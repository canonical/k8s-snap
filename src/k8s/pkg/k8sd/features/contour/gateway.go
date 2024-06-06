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

	return nil
}
