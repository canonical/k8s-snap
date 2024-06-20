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

// waitForRequiredContourCommonCRDs waits for the required contour CRDs to be available
// by checking the API resources by group version
func waitForRequiredContourCommonCRDs(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return control.WaitUntilReady(ctx, func() (bool, error) {
		resourcesV1Alpha, err := client.ListResourcesForGroupVersion("projectcontour.io/v1alpha1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}
		resourcesV1, err := client.ListResourcesForGroupVersion("projectcontour.io/v1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}
		combinedAPIResources := append(resourcesV1Alpha.APIResources, resourcesV1.APIResources...)

		requiredCRDs := map[string]bool{
			"contourconfigurations":     true,
			"contourdeployments":        true,
			"extensionservices":         true,
			"tlscertificatedelegations": true,
			"httpproxies":               true,
		}

		requiredCount := len(requiredCRDs)
		for _, resource := range combinedAPIResources {
			fmt.Printf("%s", resource.Name)
			if _, exists := requiredCRDs[resource.Name]; exists {
				requiredCount--
			}
		}
		return requiredCount == 0, nil
	})
}
