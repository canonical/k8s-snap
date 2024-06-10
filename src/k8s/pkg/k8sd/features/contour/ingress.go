package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

// ApplyIngress will install the contour helm chart when ingress.Enabled is true.
// ApplyIngress will deinstall the contour helm chart when ingress.Disabled is false.
// ApplyIngress will rollout restart the Contour pods in case any Contour configuration was changed.
// ApplyIngress returns an error if anything fails.
func ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	//TODO: map these friends
	// enableProxyProtocol = ingress.GetEnableProxyProtocol()

	if !ingress.GetEnabled() {
		if _, err := m.Apply(ctx, chartContour, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall ingress: %w", err)
		}
	}
	var values map[string]any
	if ingress.GetEnabled() {
		values = map[string]any{
			"envoy-service-namespace": "projectcontour", //TODO: Can we remove this?
			"envoy-service-name":      "envoy",
		}
	}

	changed, err := m.Apply(ctx, chartContour, helm.StatePresent, values)
	if err != nil {
		return fmt.Errorf("failed to enable ingress: %w", err)
	}
	if !changed || !ingress.GetEnabled() {
		return nil
	}

	if err := rolloutRestartContour(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart contour to apply ingress: %w", err)
	}

	// Install the delegation resource for the default TLS secret.
	// The default TLS secret is created by the user created,
	// and gets set via k8s set defaultTLSSecret=bananas.
	if ingress.GetDefaultTLSSecret() != "" {
		values = map[string]any{
			"defaultTLSSecret": ingress.GetDefaultTLSSecret(),
		}
		if _, err := m.Apply(ctx, chartDefaultTLS, helm.StatePresent, values); err != nil {
			return fmt.Errorf("failed to install the delegation resource for default TLS secret: %w", err)
		}
	} else {
		if _, err := m.Apply(ctx, chartDefaultTLS, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall the delegation resource for default TLS secret: %w", err)
		}
	}

	return nil
}

func rolloutRestartContour(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "ck-ingress-contour-envoy", "projectcontour"); err != nil { //TODO: check name of deployment
			return fmt.Errorf("failed to restart contour deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDaemonset(ctx, "ck-ingress-contour-envoy", "projectcontour"); err != nil {
			return fmt.Errorf("failed to restart contour daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour daemonset after %d attempts: %w", attempts, err)
	}

	return nil
}
