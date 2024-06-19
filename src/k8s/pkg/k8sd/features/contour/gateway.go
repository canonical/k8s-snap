package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
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
		return nil
	}

	// Apply common contour CRDS, these are shared with ingress
	if err := applyCommonContourCRDS(ctx, snap, true); err != nil {
		return fmt.Errorf("failed to apply common contour CRDS: %w", err)
	}

	if err := waitForRequiredContourCommonCRDs(ctx, snap); err != nil {
		return fmt.Errorf("failed to wait for required contour common CRDs to be available: %w", err)
	}

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, nil); err != nil {
		return fmt.Errorf("failed to install the contour gateway chart: %w", err)
	}

	return nil
}

func waitForRequiredContourCommonCRDs(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	requiredCRDs := map[string][]string{
		"projectcontour.io/v1alpha1": {
			"contourconfigurations.projectcontour.io",
			"contourdeployments.projectcontour.io",
			"extensionservices.projectcontour.io",
		},
		"projectcontour.io/v1": {
			"tlscertificatedelegations.projectcontour.io",
			"httpproxies.projectcontour.io",
		},
	}

	// checkRequiredCRDs checks if the required CRDs are present in the cluster.
	checkRequiredCRDs := func(groupVersion string, required []string) (bool, error) {
		resources, err := client.ListResourcesForGroupVersion(groupVersion)
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}

		requiredMap := make(map[string]bool)
		for _, crd := range required {
			requiredMap[crd] = true
		}

		for _, resource := range resources.APIResources {
			delete(requiredMap, resource.Name)
		}

		return len(requiredMap) == 0, nil
	}

	return control.WaitUntilReady(ctx, func() (bool, error) {
		for groupVersion, crds := range requiredCRDs {
			ready, err := checkRequiredCRDs(groupVersion, crds)
			if err != nil {
				return false, err
			}
			if !ready {
				return false, nil
			}
		}
		return true, nil
	})
}
