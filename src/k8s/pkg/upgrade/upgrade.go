package upgrade

import (
	"fmt"

	"github.com/canonical/k8s/pkg/version"
)

// GetName returns the name of the upgrade resource based on the version info.
func GetName(v version.Info) string {
	return fmt.Sprintf("cluster-upgrade-to-k8s-%s-rev-%s", v.KubernetesVersion, v.Revision)
}
