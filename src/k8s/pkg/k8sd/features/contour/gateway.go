package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyGateway will install a helm chart for contour-gateway-provisioner on the cluster when gateway.Enabled is true.
// ApplyGateway will uninstall the helm chart for contour-gateway-provisioner from the cluster when gateway.Enabled is false.
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
