package features

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

func ApplyLocalStorage(ctx context.Context, snap snap.Snap, cfg types.LocalStorage) error {
	m := newHelm(snap)

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

	_, err := m.Apply(ctx, featureLocalStorage, stateFromBool(cfg.GetEnabled()), values)
	return err
}
