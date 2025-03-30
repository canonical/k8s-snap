package metrics_server

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

const (
	enabledMsg          = "enabled"
	disabledMsg         = "disabled"
	deleteFailedMsgTmpl = "Failed to delete Metrics Server, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy Metrics Server, the error was: %v"
)

const METRICS_SERVER_VERSION = "v1.0.0"

// ApplyMetricsServer deploys metrics-server when cfg.Enabled is true.
// ApplyMetricsServer removes metrics-server when cfg.Enabled is false.
// ApplyMetricsServer will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyMetricsServer returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyMetricsServer(ctx context.Context, _ state.State, snap snap.Snap, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	values := metricsServerValues{}

	if err := values.applyDefaults(); err != nil {
		err = fmt.Errorf("failed to apply defaults: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyImages(); err != nil {
		err = fmt.Errorf("failed to apply images: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyAnnotations(annotations); err != nil {
		err = fmt.Errorf("failed to apply annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	_, err := m.Apply(ctx, features.MetricsServer, METRICS_SERVER_VERSION, chart, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	if err != nil {
		if cfg.GetEnabled() {
			err = fmt.Errorf("failed to install metrics server chart: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: imageTag,
				Message: fmt.Sprintf(deployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to delete metrics server chart: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: imageTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
	} else {
		if cfg.GetEnabled() {
			return types.FeatureStatus{
				Enabled: true,
				Version: imageTag,
				Message: enabledMsg,
			}, nil
		} else {
			return types.FeatureStatus{
				Enabled: false,
				Version: imageTag,
				Message: disabledMsg,
			}, nil
		}
	}
}
