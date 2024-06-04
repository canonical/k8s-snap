package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

// ApplyIngress assumes that the managed Contour CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyIngress will enable Contour's ingress controller when ingress.Enabled is true.
// ApplyIngress will disable Contour's ingress controller when ingress.Disabled is false.
// ApplyIngress will rollout restart the Contour pods in case any Contour configuration was changed.
// ApplyIngress returns an error if anything fails.
func ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	var values map[string]any
	if ingress.GetEnabled() {
		values = map[string]any{
			"envoy-service-namespace": "project-contour",
			"envoy-service-name":      "envoy",
			"tls": map[string]any{
				"envoy-client-certificate": ingress.GetDefaultTLSSecret(), //TODO: I think this is wrong
			},
			"gateway": map[string]any{
				"gatewayRef": map[string]any{
					"name":      "gateway",
					"namespace": "project-contour",
				},
			},
		}
	}

	changed, err := m.Apply(ctx, chartContour, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values)
	if err != nil {
		return fmt.Errorf("failed to enable ingress: %w", err)
	}
	if !changed || !ingress.GetEnabled() {
		return nil
	}

	if err := rolloutRestartContour(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart contour to apply ingress: %w", err)
	}
	return nil
}

func rolloutRestartContour(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "contour-contour", "project-contour"); err != nil { //TODO: check name of deployment
			return fmt.Errorf("failed to restart contour deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDaemonset(ctx, "contour-envoy", "project-contour"); err != nil {
			return fmt.Errorf("failed to restart contour daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour daemonset after %d attempts: %w", attempts, err)
	}

	return nil
}
