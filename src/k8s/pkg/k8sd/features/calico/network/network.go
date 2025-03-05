package network

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/calico"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	deployFailedMsgTmpl = "Failed to deploy Calico, the error was: %v"
	deleteFailedMsgTmpl = "Failed to delete Calico, the error was: %v"
)

// ApplyNetwork will deploy Calico when network.Enabled is true.
// ApplyNetwork will remove Calico when network.Enabled is false.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r reconciler) Reconcile(ctx context.Context, cfg types.ClusterConfig) (types.FeatureStatus, error) {
	calicoImage := r.Manifest().GetImage(CalicoImageName)

	helmClient := r.HelmClient()

	network := cfg.Network
	annotations := cfg.Annotations

	if !network.GetEnabled() {
		if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(CalicoChartName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall network: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: calicoImage.Tag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}

		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: calico.DisabledMsg,
		}, nil
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		err = fmt.Errorf("failed to apply default values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyImageOverrides(r.Manifest()); err != nil {
		err = fmt.Errorf("failed to calculate image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfiguration(network); err != nil {
		err = fmt.Errorf("failed to calculate cluster config values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyAnnotations(annotations); err != nil {
		err = fmt.Errorf("failed to calculate annotation overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(CalicoChartName), helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to enable network: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: calicoImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: calicoImage.Tag,
		Message: calico.EnabledMsg,
	}, nil
}
