package network

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	ciliumAgentImage := FeatureNetwork.GetImage(CiliumAgentImageName)
	ciliumOperatorImage := FeatureNetwork.GetImage(CiliumOperatorImageName)

	images.Register(
		fmt.Sprintf("%s:%s", ciliumAgentImage.GetURI(), ciliumAgentImage.Tag),
		fmt.Sprintf("%s-generic:%s", ciliumOperatorImage.GetURI(), ciliumOperatorImage.Tag),
	)
}
