package setup

import (
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

var k8sServices = []string{"containerd", "k8s-dqlite", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "kubelet"}

// InitServiceArgs handles the setup of services arguments.
//   - For each service, copies the default arguments files from the snap under $SNAP_DATA/args and apply any overwrites
//   - Note that the `k8sd` service is already configured in the snap install hook and thus not included here
func InitServiceArgs(snap snap.Snap, extraArgs map[string]map[string]string) error {
	for _, service := range k8sServices {
		serviceArgs, err := utils.ParseArgumentFile(snap.Path("k8s/args", service))
		if err != nil {
			return fmt.Errorf("failed to parse argument file for %s: %w", service, err)
		}

		// Apply overwrites for each service
		if args, exists := extraArgs[service]; exists {
			for argument, value := range args {
				serviceArgs[argument] = value
			}
		}

		if err := utils.SerializeArgumentFile(serviceArgs, snap.DataPath("args", service)); err != nil {
			return fmt.Errorf("failed to write arguments file for %s: %w", service, err)
		}
	}

	return nil
}
