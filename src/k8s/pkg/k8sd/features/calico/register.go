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
		fmt.Sprintf("%s/apiserver:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/cni:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/csi:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/ctl:%s", calicoImageRepo, calicoCtlTag),
		fmt.Sprintf("%s/kube-controllers:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/node:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/node-driver-registrar:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/pod2daemon-flexvol:%s", calicoImageRepo, calicoTag),
		fmt.Sprintf("%s/typha:%s", calicoImageRepo, calicoTag),
	)
}
