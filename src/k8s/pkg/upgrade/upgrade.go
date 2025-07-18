package upgrade

import (
	"fmt"

	"github.com/canonical/k8s/pkg/version"
)

// GetName returns the name of the upgrade resource based on the version info.
func GetName(v version.Info) string {
	return fmt.Sprintf("cluster-upgrade-to-rev-%s", v.Revision)
}
