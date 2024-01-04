package setup

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

var k8sServices = []string{"containerd", "k8s-dqlite", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "kubelet"}

// InitServiceArgs handles the setup of services arguments.
//   - For each service, copies the default arguments files from the snap under $SNAP_DATA/args and apply any overwrites
//   - Note that the `k8sd` service is already configured in the snap install hook and thus not included here
func InitServiceArgs(snap snap.Snap, overwrites apiv1.ExtraServiceArgs) error {
	for _, service := range k8sServices {
		serviceArgs, err := utils.ParseArgumentFile(snap.Path("k8s/args", service))
		if err != nil {
			return fmt.Errorf("failed to parse argument file for %s: %w", service, err)
		}

		// Apply overwrites for each service
		if ao, exists := overwrites[service]; exists {
			for argument, value := range ao {
				serviceArgs[argument] = value
			}
		}

		err = utils.SerializeArgumentFile(serviceArgs, snap.DataPath("args", service))
	}

	return nil
}
