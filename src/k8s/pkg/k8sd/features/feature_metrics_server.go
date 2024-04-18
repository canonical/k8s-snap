package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

func ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer) error {
	m := newHelm(snap)

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

	_, err := m.Apply(ctx, featureMetricsServer, stateFromBool(cfg.GetEnabled()), values)
	return err
}
