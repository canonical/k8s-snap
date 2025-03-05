package network

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	ciliumAgentImage := Manifest.GetImage(CiliumAgentImageName)
	ciliumOperatorImage := Manifest.GetImage(CiliumOperatorImageName)

	images.Register(
		fmt.Sprintf("%s:%s", ciliumAgentImage.GetURI(), ciliumAgentImage.Tag),
		fmt.Sprintf("%s-generic:%s", ciliumOperatorImage.GetURI(), ciliumOperatorImage.Tag),
	)

	features.Register(&Manifest)
}
