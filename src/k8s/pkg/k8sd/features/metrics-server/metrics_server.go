package metrics_server

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	enabledMsg          = "enabled"
	disabledMsg         = "disabled"
	deleteFailedMsgTmpl = "Failed to delete Metrics Server, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy Metrics Server, the error was: %v"
)

// ApplyMetricsServer deploys metrics-server when cfg.Enabled is true.
// ApplyMetricsServer removes metrics-server when cfg.Enabled is false.
// ApplyMetricsServer will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyMetricsServer returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyMetricsServer(ctx context.Context, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	status := types.FeatureStatus{
		Version: imageTag,
		Enabled: cfg.GetEnabled(),
	}
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
	if err != nil {
		if cfg.GetEnabled() {
			enableErr := fmt.Errorf("failed to install metrics server chart: %w", err)
			status.Message = fmt.Sprintf(deployFailedMsgTmpl, enableErr)
			status.Enabled = false
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to delete metrics server chart: %w", err)
			status.Message = fmt.Sprintf(deleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	} else {
		if cfg.GetEnabled() {
			status.Message = enabledMsg
			return status, nil
		} else {
			status.Message = disabledMsg
			return status, nil
		}
	}
}
