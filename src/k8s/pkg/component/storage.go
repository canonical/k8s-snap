package component

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

func EnableStorageComponent(s snap.Snap, ctx context.Context) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	var values map[string]any = nil
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
