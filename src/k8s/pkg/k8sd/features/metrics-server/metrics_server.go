package metrics_server

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyMetricsServer deploys metrics-server when cfg.Enabled is true.
// ApplyMetricsServer removes metrics-server when cfg.Enabled is false.
// ApplyMetricsServer returns an error if anything fails.
func ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) error {
	m := snap.HelmClient()

	config := internalConfig(annotations)

	values := map[string]any{
		"image": map[string]any{
			"repository": config.imageRepo,
			"tag":        config.imageTag,
		},
		"securityContext": map[string]any{
			// ROCKs with Pebble as the entrypoint do not work with this option.
			"readOnlyRootFilesystem": false,
		},
	}

	_, err := m.Apply(ctx, chart, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	return err
}
