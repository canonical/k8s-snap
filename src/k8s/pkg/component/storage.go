package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/vals"
)

func UpdateStorageComponent(ctx context.Context, s snap.Snap, isRefresh bool, config types.LocalStorage) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	values := map[string]any{
		"storageClass": map[string]any{
			"enabled":       true,
			"isDefault":     config.GetDefault(),
			"reclaimPolicy": config.GetReclaimPolicy(),
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
				"path": config.GetLocalPath(),
			},
		},
	}

	if isRefresh {
		if err := manager.Refresh("local-storage", values); err != nil {
			return fmt.Errorf("failed to enable local-storage component: %w", err)
		}
	} else {
		if err := manager.Enable("local-storage", values); err != nil {
			return fmt.Errorf("failed to enable local-storage component: %w", err)
		}
	}

	return nil
}

func DisableStorageComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("local-storage"); err != nil {
		return fmt.Errorf("failed to disable local-storage component: %w", err)
	}

	return nil
}

func ReconcileLocalStorageComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateStorageComponent(ctx, s, true, clusterConfig.LocalStorage)
		if err != nil {
			return fmt.Errorf("failed to refresh local-storage: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateStorageComponent(ctx, s, false, clusterConfig.LocalStorage)
		if err != nil {
			return fmt.Errorf("failed to enable local-storage: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableStorageComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable local-storage: %w", err)
		}
		return nil
	}
	return nil
}
