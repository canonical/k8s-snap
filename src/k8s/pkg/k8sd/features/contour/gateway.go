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

	values := map[string]any{
		"projectcontour": map[string]any{
			"image": map[string]any{
				"repository": contourGatewayImageRepo,
				"tag":        contourGatewayImageTag,
			},
		},
		"envoyproxy": map[string]any{
			"image": map[string]any{
				"repository": envoyProxyImageRepo,
				"tag":        envoyProxyImageTag,
			},
		},
	}

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, values); err != nil {
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

		requiredCRDs := map[string]bool{
			"projectcontour.io/v1alpha1:contourconfigurations": true,
			"projectcontour.io/v1alpha1:contourdeployments":    true,
			"projectcontour.io/v1alpha1:extensionservices":     true,
			"projectcontour.io/v1:tlscertificatedelegations":   true,
			"projectcontour.io/v1:httpproxies":                 true,
		}

		requiredCount := len(requiredCRDs)
		for _, resource := range resourcesV1Alpha.APIResources {
			if _, exists := requiredCRDs[fmt.Sprintf("projectcontour.io/v1alpha1:%s", resource.Name)]; exists {
				requiredCount--
			}
		}
		for _, resource := range resourcesV1.APIResources {
			if _, exists := requiredCRDs[fmt.Sprintf("projectcontour.io/v1:%s", resource.Name)]; exists {
				requiredCount--
			}
		}

		return requiredCount == 0, nil
	})
}
