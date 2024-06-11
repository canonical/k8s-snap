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
// ApplyGateway will apply common contour CRDS, these are shared with ingress.
func ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	if !gateway.GetEnabled() {
		if _, err := m.Apply(ctx, chartGateway, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall the contour gateway chart: %w", err)
		}
	}

	// Apply common contour CRDS, these are shared with ingress
	if err := applyCommonContourCRDS(ctx, snap, true); err != nil { //TODO: check wether one of ingress/gateway is enabled
		return fmt.Errorf("failed to apply common contour CRDS: %w", err)
	}

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, nil); err != nil {
		return fmt.Errorf("failed to install the contour gateway chart: %w", err)
	}

	return nil
}
