package loadbalancer

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	metalLBControllerImage := Manifest.GetImage(MetalLBControllerImageName)
	metalLBSpeakerImage := Manifest.GetImage(MetalLBSpeakerImageName)
	frrImage := Manifest.GetImage(FRRImageName)

	images.Register(
		fmt.Sprintf("%s:%s", metalLBControllerImage.GetURI(), metalLBControllerImage.Tag),
		fmt.Sprintf("%s:%s", metalLBSpeakerImage.GetURI(), metalLBSpeakerImage.Tag),
		fmt.Sprintf("%s:%s", frrImage.GetURI(), frrImage.Tag),
	)

	features.Register(&Manifest)
}
