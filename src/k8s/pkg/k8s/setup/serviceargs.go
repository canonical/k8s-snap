package setup

import (
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8s/utils"
)

// InitServiceArgs handles the setup of services arguments.
//   - For each service, copies the default arguments files from the snap under $SNAP_DATA/args
func InitServiceArgs() error {
	for _, service := range []string{"containerd", "k8sd", "k8s-dqlite", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "kubelet"} {
		err := utils.CopyFile(filepath.Join(utils.SNAP, "k8s/args", service), filepath.Join(utils.SNAP_DATA, "args", service))
		if err != nil {
			return fmt.Errorf("failed to copy %s args: %w", service, err)
		}
	}

	return nil
}
