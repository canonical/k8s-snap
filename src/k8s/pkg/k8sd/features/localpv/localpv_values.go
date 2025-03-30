package localpv

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type localPVValues map[string]any

func (v *localPVValues) applyDefaults() error {
	values := localPVValues{
		"storageClass": map[string]any{
			"enabled": true,
		},
		"serviceMonitor": map[string]any{
			"enabled": false,
		},
		"controller": map[string]any{
			"csiDriverArgs": []string{"--args", "rawfile", "csi-driver", "--disable-metrics"},
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *localPVValues) applyImages() error {
	values := localPVValues{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        ImageTag,
			},
		},
		"node": map[string]any{
			"image": map[string]any{
				"repository": imageRepo,
				"tag":        ImageTag,
			},
		},
		"images": map[string]any{
			"csiNodeDriverRegistrar": csiNodeDriverImage,
			"csiProvisioner":         csiProvisionerImage,
			"csiResizer":             csiResizerImage,
			"csiSnapshotter":         csiSnapshotterImage,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *localPVValues) applyClusterConfig(storage types.LocalStorage) error {
	values := localPVValues{
		"storageClass": map[string]any{
			"isDefault":     storage.GetDefault(),
			"reclaimPolicy": storage.GetReclaimPolicy(),
		},
		"node": map[string]any{
			"storage": map[string]any{
				"path": storage.GetLocalPath(),
			},
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
