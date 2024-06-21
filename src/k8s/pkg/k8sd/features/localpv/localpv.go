package localpv

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyLocalStorage deploys the rawfile-localpv CSI driver on the cluster based on the given configuration, when cfg.Enabled is true.
// ApplyLocalStorage removes the rawfile-localpv when cfg.Enabled is false.
// ApplyLocalStorage returns an error if anything fails.
func ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage, _ types.Annotations) error {
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
			"csiNodeDriverRegistrar": fmt.Sprintf("%s/%s", csiImageRepo, csiNodeDriverImage),
			"csiProvisioner":         fmt.Sprintf("%s/%s", csiImageRepo, csiProvisionerImage),
			"csiResizer":             fmt.Sprintf("%s/%s", csiImageRepo, csiResizerImage),
			"csiSnapshotter":         fmt.Sprintf("%s/%s", csiImageRepo, csiSnapshotterImage),
		},
	}

	_, err := m.Apply(ctx, chart, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	return err
}
