package upgrade

import (
	"fmt"

	"github.com/canonical/k8s/pkg/version"
)

// GetName returns the name of the upgrade resource based on the version info.
func GetName(v version.Info) string {
	var k8sVersionStr string
	if v.KubernetesVersion == nil {
		k8sVersionStr = "UNKNOWN"
	} else {
		k8sVersionStr = fmt.Sprintf("%v", v.KubernetesVersion)
	}
	return fmt.Sprintf("cluster-upgrade-to-k8s-%s-rev-%s", k8sVersionStr, v.Revision)
}
