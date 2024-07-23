package localpv

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	enabledMsg          = "enabled at %s"
	disabledMsg         = "disabled"
	deployFailedMsgTmpl = "Failed to deploy Local Storage, the error was: %v"
	deleteFailedMsgTmpl = "Failed to delete Local Storage, the error was: %v"
)

// ApplyLocalStorage deploys the rawfile-localpv CSI driver on the cluster based on the given configuration, when cfg.Enabled is true.
// ApplyLocalStorage removes the rawfile-localpv when cfg.Enabled is false.
// ApplyLocalStorage will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLocalStorage returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage, _ types.Annotations) (types.FeatureStatus, error) {
	status := types.FeatureStatus{
		Version: imageTag,
		Enabled: cfg.GetEnabled(),
	}
	m := snap.HelmClient()

	values := map[string]any{
		"storageClass": map[string]any{
			"enabled":       true,
			"isDefault":     cfg.GetDefault(),
			"reclaimPolicy": cfg.GetReclaimPolicy(),
		},
		"serviceMonitor": map[string]any{
			"enabled": false,
		},
		"controller": map[string]any{
			"csiDriverArgs": []string{"--args", "rawfile", "csi-driver", "--disable-metrics"},
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        imageTag,
			},
		},
		"node": map[string]any{
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        imageTag,
			},
			"storage": map[string]any{
				"path": cfg.GetLocalPath(),
			},
		},
		"images": map[string]any{
			"csiNodeDriverRegistrar": csiNodeDriverImage,
			"csiProvisioner":         csiProvisionerImage,
			"csiResizer":             csiResizerImage,
			"csiSnapshotter":         csiSnapshotterImage,
		},
	}

	_, err := m.Apply(ctx, chart, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	if err != nil {
		if cfg.GetEnabled() {
			enableErr := fmt.Errorf("failed to install rawfile-csi helm package: %w", err)
			status.Message = fmt.Sprintf(deployFailedMsgTmpl, enableErr)
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to delete rawfile-csi helm package: %w", err)
			status.Message = fmt.Sprintf(deleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	} else {
		if cfg.GetEnabled() {
			status.Message = fmt.Sprintf(enabledMsg, cfg.GetLocalPath())
			return status, nil
		} else {
			status.Version = ""
			status.Message = disabledMsg
			return status, nil
		}
	}
}
