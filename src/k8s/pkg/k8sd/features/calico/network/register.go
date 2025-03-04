package network

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features/manifests"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	tigeraOperatorImage := FeatureNetwork.GetImage(TigeraOperatorImageName)
	calicoCtlImage := FeatureNetwork.GetImage(CalicoCtlImageName)
	calicoImage := FeatureNetwork.GetImage(CalicoImageName)

	images.Register(
		// Tigera images
		fmt.Sprintf("%s/%s:%s", tigeraOperatorImage.Registry, tigeraOperatorImage.Repository, tigeraOperatorImage.Tag),
		// Calico images
		fmt.Sprintf("%s/apiserver:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/cni:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/csi:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/ctl:%s", calicoCtlImage.GetURI(), calicoCtlImage.Tag),
		fmt.Sprintf("%s/kube-controllers:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/node:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/node-driver-registrar:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/pod2daemon-flexvol:%s", calicoImage.GetURI(), calicoImage.Tag),
		fmt.Sprintf("%s/typha:%s", calicoImage.GetURI(), calicoImage.Tag),
	)

	manifests.Register(&manifest)
}
