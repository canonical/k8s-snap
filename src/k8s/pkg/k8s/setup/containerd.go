package setup

import (
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
)

// InitContainerd handles the setup of containerd.
//   - Copies required files and binaries needed by Containerd to the correct paths.
func InitContainerd() error {
	err := utils.CopyFile(utils.Path("k8s/config/containerd/config.toml"), "/etc/containerd/config.toml")
	if err != nil {
		return fmt.Errorf("failed to copy containerd config: %w", err)
	}

	err = utils.CopyDirectory(utils.Path("opt/cni/bin/"), "/opt/cni/bin")
	if err != nil {
		return fmt.Errorf("failed to copy cni/bin: %w", err)
	}

	err = utils.ChmodRecursive("/opt/cni/bin", 0700)
	if err != nil {
		return fmt.Errorf("failed to adjust permissions of /opt/cni/bin: %w", err)
	}

	return nil
}
