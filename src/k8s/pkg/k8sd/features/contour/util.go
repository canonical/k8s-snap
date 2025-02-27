package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

// WaitForRequiredContourCommonCRDs waits for the required contour CRDs to be available
// by checking the API resources by group version.
func WaitForRequiredContourCommonCRDs(ctx context.Context, snap snap.Snap) error {
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

// RolloutRestartContour will rollout restart the Contour pods in case any Contour configuration was changed.
func RolloutRestartContour(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "ck-ingress-contour-contour", "projectcontour"); err != nil {
			return fmt.Errorf("failed to restart contour deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour deployment after %d attempts: %w", attempts, err)
	}

	return nil
}
