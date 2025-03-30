package localpv

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
	enabledMsg          = "enabled at %s"
	disabledMsg         = "disabled"
	deployFailedMsgTmpl = "Failed to deploy Local Storage, the error was: %v"
	deleteFailedMsgTmpl = "Failed to delete Local Storage, the error was: %v"
)

const LOCAL_STORAGE_VERSION = "v1.0.0"

// ApplyLocalStorage deploys the rawfile-localpv CSI driver on the cluster based on the given configuration, when cfg.Enabled is true.
// ApplyLocalStorage removes the rawfile-localpv when cfg.Enabled is false.
// ApplyLocalStorage will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLocalStorage returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyLocalStorage(ctx context.Context, _ state.State, snap snap.Snap, cfg types.LocalStorage, _ types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	values := localPVValues{}

	if err := values.applyDefaults(); err != nil {
		err = fmt.Errorf("failed to apply defaults: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyImages(); err != nil {
		err = fmt.Errorf("failed to apply images: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfig(cfg); err != nil {
		err = fmt.Errorf("failed to apply cluster config: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if _, err := m.Apply(ctx, features.LocalStorage, LOCAL_STORAGE_VERSION, Chart, helm.StatePresentOrDeleted(cfg.GetEnabled()), values); err != nil {
		if cfg.GetEnabled() {
			err = fmt.Errorf("failed to install rawfile-csi helm package: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ImageTag,
				Message: fmt.Sprintf(deployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to delete rawfile-csi helm package: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ImageTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
	}

	if cfg.GetEnabled() {
		return types.FeatureStatus{
			Enabled: true,
			Version: ImageTag,
			Message: fmt.Sprintf(enabledMsg, cfg.GetLocalPath()),
		}, nil
	} else {
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: disabledMsg,
		}, nil
	}
}
