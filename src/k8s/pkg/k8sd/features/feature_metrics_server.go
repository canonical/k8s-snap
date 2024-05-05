package features

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyMetricsServer is used to configure the metrics-server feature on Canonical Kubernetes.
// ApplyMetricsServer deploys metrics-server when cfg.Enabled is true.
// ApplyMetricsServer removes metrics-server when cfg.Enabled is false.
// ApplyMetricsServer returns an error if anything fails.
func ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer) error {
	m := snap.HelmClient()

	values := map[string]any{
		"image": map[string]any{
			"repository": metricsServerImageRepository,
			"tag":        metricsServerImageTag,
		},
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	_, err := m.Apply(ctx, chartMetricsServer, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	return err
}
