package setup

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
)

// InitPermissions makes sure(sets up) the permissions of paths utilized by the snap are correct.
func InitPermissions(ctx context.Context) error {
	// Shelling out since go doesn't support symbolic mode definitions.
	err := utils.RunCommand(ctx,
		"chmod", "go-rxw", "-R",
		utils.SnapDataPath("args"),
		utils.SnapCommonPath("opt"),
		utils.SnapCommonPath("etc"),
		utils.SnapCommonPath("var/lib"),
		utils.SnapCommonPath("var/log"),
	)
	if err != nil {
		return fmt.Errorf("failed to change folder permissions: %w", err)
	}

	return nil
}
