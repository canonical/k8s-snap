package loadbalancer

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	metalLBControllerImage := FeatureLoadBalancer.GetImage(MetalLBControllerImageName)
	metalLBSpeakerImage := FeatureLoadBalancer.GetImage(MetalLBSpeakerImageName)
	frrImage := FeatureLoadBalancer.GetImage(FRRImageName)

	images.Register(
		fmt.Sprintf("%s:%s", metalLBControllerImage.GetURI(), metalLBControllerImage.Tag),
		fmt.Sprintf("%s:%s", metalLBSpeakerImage.GetURI(), metalLBSpeakerImage.Tag),
		fmt.Sprintf("%s:%s", frrImage.GetURI(), frrImage.Tag),
	)
}
