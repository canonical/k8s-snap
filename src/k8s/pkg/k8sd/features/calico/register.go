package calico

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		// Tigera images
		fmt.Sprintf("%s/%s:%s", imageRepo, tigeraOperatorImage, tigeraOperatorVersion),
		// Calico images
		fmt.Sprintf("%s/apiserver:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/cni:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/csi:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/ctl:%s", calicoImageRepo, calicoCtlTag),
		fmt.Sprintf("%s/kube-controllers:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/node:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/node-driver-registrar:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/pod2daemon-flexvol:%s", calicoImageRepo, CalicoTag),
		fmt.Sprintf("%s/typha:%s", calicoImageRepo, CalicoTag),
	)
}
