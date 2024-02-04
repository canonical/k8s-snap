package setup

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

// InitPermissions makes sure(sets up) the permissions of paths utilized by the snap are correct.
func InitPermissions(ctx context.Context, snap snap.Snap) error {
	// Shelling out since go doesn't support symbolic mode definitions.
	err := utils.RunCommand(ctx,
		"chmod", "go-rxw", "-R",
		snap.DataPath("args"),
		snap.CommonPath("etc"),
		snap.CommonPath("var/lib"),
		"/opt/cni/bin",
	)
	if err != nil {
		return fmt.Errorf("failed to change folder permissions: %w", err)
	}

	return nil
}
