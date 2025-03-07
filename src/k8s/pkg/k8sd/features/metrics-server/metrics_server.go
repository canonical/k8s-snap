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
func ApplyMetricsServer(ctx context.Context, snap snap.Snap, m helm.Client, cfg types.MetricsServer, annotations types.Annotations) (types.FeatureStatus, error) {
	metricsServerImage := FeatureMetricsServer.GetImage(MetricsServerImageName)
	imageTag := metricsServerImage.Tag

	config := config{}

	if config.imageTag != "" {
		imageTag = config.imageTag
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		err = fmt.Errorf("failed to apply default values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyAnnotations(annotations); err != nil {
		err = fmt.Errorf("failed to apply annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: imageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	_, err := m.Apply(ctx, FeatureMetricsServer.GetChart(MetricsServerChartName), helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
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
