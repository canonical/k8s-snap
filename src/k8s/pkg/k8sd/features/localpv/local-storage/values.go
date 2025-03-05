package local_storage

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
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

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyImageOverrides(manifest types.FeatureManifest) error {
	rawFileImage := manifest.GetImage(RawFileImageName)
	csiNodeDriverImage := manifest.GetImage(CSINodeDriverImageName)
	csiProvisionerImage := manifest.GetImage(CSIProvisionerImageName)
	csiResizerImage := manifest.GetImage(CSIResizerImageName)
	csiSnapshotterImage := manifest.GetImage(CSISnapshotterImageName)

	values := map[string]any{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": rawFileImage.GetURI(),
				"tag":        rawFileImage.Tag,
			},
		},
		"node": map[string]any{
			"image": map[string]any{
				"repository": rawFileImage.GetURI(),
				"tag":        rawFileImage.Tag,
			},
		},
		"images": map[string]any{
			"csiNodeDriverRegistrar": fmt.Sprintf("%s:%s", csiNodeDriverImage.GetURI(), csiNodeDriverImage.Tag),
			"csiProvisioner":         fmt.Sprintf("%s:%s", csiProvisionerImage.GetURI(), csiProvisionerImage.Tag),
			"csiResizer":             fmt.Sprintf("%s:%s", csiResizerImage.GetURI(), csiResizerImage.Tag),
			"csiSnapshotter":         fmt.Sprintf("%s:%s", csiSnapshotterImage.GetURI(), csiSnapshotterImage.Tag),
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(cfg types.LocalStorage) error {
	values := map[string]any{
		"storageClass": map[string]any{
			"isDefault":     cfg.GetDefault(),
			"reclaimPolicy": cfg.GetReclaimPolicy(),
		},
		"node": map[string]any{
			"storage": map[string]any{
				"path": cfg.GetLocalPath(),
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration values: %w", err)
	}

	return nil
}
