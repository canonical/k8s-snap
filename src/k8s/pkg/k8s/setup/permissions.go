package setup

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8s/utils"
)

// InitPermissions makes sure(sets up) the permissions of paths utilized by the snap are correct.
func InitPermissions() error {
	// Shelling out since go doesn't support symbolic mode definitions.
	chmcmd := exec.Command("chmod", "go-rxw", "-R", filepath.Join(utils.SNAP_DATA, "args"), filepath.Join(utils.SNAP_COMMON, "opt"), filepath.Join(utils.SNAP_COMMON, "etc"), filepath.Join(utils.SNAP_COMMON, "var/lib"), filepath.Join(utils.SNAP_COMMON, "var/log"))

	_, err := chmcmd.Output()
	if err != nil {
		return fmt.Errorf("failed to change folder permissions: %w", err)
	}

	return nil
}
