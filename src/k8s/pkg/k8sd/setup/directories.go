package setup

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/snap"
)

// EnsureAllDirectories ensures all required configuration and state directories are created.
func EnsureAllDirectories(snap snap.Snap) error {
	for _, dir := range []string{
		snap.CNIBinDir(),
		snap.CNIConfDir(),
		snap.ContainerdConfigDir(),
		snap.ContainerdExtraConfigDir(),
		snap.ContainerdRegistryConfigDir(),
		snap.ContainerdRootDir(),
		snap.ContainerdSocketDir(),
		snap.ContainerdStateDir(),
		snap.K8sDqliteStateDir(),
		snap.K8sdStateDir(),
		snap.KubernetesConfigDir(),
		snap.KubernetesPKIDir(),
		snap.KubeletRootDir(),
		snap.ServiceArgumentsDir(),
		snap.ServiceExtraConfigDir(),
	} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create required directory: %w", err)
		}
	}
	return nil
}
