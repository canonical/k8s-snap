package local_storage

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	deployFailedMsgTmpl = "Failed to deploy Local Storage, the error was: %v"
	deleteFailedMsgTmpl = "Failed to delete Local Storage, the error was: %v"
)

// ApplyLocalStorage deploys the rawfile-localpv CSI driver on the cluster based on the given configuration, when cfg.Enabled is true.
// ApplyLocalStorage removes the rawfile-localpv when cfg.Enabled is false.
// ApplyLocalStorage will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLocalStorage returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r LocalStorageReconciler) ApplyLocalStorage(ctx context.Context, cfg types.LocalStorage, _ types.Annotations) (types.FeatureStatus, error) {
	rawFileImage := FeatureLocalStorage.GetImage(RawFileImageName)

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		err = fmt.Errorf("failed to apply default values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: rawFileImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: rawFileImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfiguration(cfg); err != nil {
		err = fmt.Errorf("failed to apply cluster configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: rawFileImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if _, err := r.HelmClient().Apply(ctx, FeatureLocalStorage.GetChart(RawFileChartName), helm.StatePresentOrDeleted(cfg.GetEnabled()), values); err != nil {
		if cfg.GetEnabled() {
			err = fmt.Errorf("failed to install rawfile-csi helm package: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: rawFileImage.Tag,
				Message: fmt.Sprintf(deployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to delete rawfile-csi helm package: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: rawFileImage.Tag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
	}

	if cfg.GetEnabled() {
		return types.FeatureStatus{
			Enabled: true,
			Version: rawFileImage.Tag,
			Message: fmt.Sprintf(localpv.EnabledMsg, cfg.GetLocalPath()),
		}, nil
	} else {
		return types.FeatureStatus{
			Enabled: false,
			Version: rawFileImage.Tag,
			Message: localpv.DisabledMsg,
		}, nil
	}
}
