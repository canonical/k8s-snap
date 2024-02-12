package component

import (
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

func EnableStorageComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	values := map[string]any{
		"storageClass": map[string]any{
			"enabled": true,
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
				"path": "/var/snap/k8s/common/rawfile-storage",
			},
		},
	}

	err = manager.Enable("storage", values)
	if err != nil {
		return fmt.Errorf("failed to enable storage component: %w", err)
	}

	return nil
}

func DisableStorageComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	err = manager.Disable("storage")
	if err != nil {
		return fmt.Errorf("failed to disable storage component: %w", err)
	}

	return nil
}
