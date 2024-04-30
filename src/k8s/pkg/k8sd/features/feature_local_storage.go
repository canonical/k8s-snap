package features

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyLocalStorage is used to configure the Local Storage feature on Canonical Kubernetes.
// ApplyLocalStorage deploys the rawfile-localpv CSI driver on the cluster based on the given configuration, when cfg.Enabled is true.
// ApplyLocalStorage removes the rawfile-localpv when cfg.Enabled is false.
// ApplyLocalStorage returns an error if anything fails.
func ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage) error {
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
				"repository": storageImageRepository,
				"tag":        storageImageTag,
			},
		},
		"node": map[string]any{
			"image": map[string]any{
				"repository": storageImageRepository,
				"tag":        storageImageTag,
			},
			"storage": map[string]any{
				"path": cfg.GetLocalPath(),
			},
		},
	}

	_, err := m.Apply(ctx, chartLocalStorage, helm.StatePresentOrDeleted(cfg.GetEnabled()), values)
	return err
}
