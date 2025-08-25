package metallb

import (
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		MetalLBControllerImage().String(),
		MetalLBSpeakerImage().String(),
		FRRImage().String(),
	)
}
