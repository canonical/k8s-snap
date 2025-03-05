package ingress

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	contourIngressContourImage := FeatureIngress.GetImage(ContourIngressContourImageName)
	contourIngressEnvoyImage := FeatureIngress.GetImage(ContourIngressEnvoyImageName)

	images.Register(
		fmt.Sprintf("%s:%s", contourIngressEnvoyImage.GetURI(), contourIngressEnvoyImage.Tag),
		fmt.Sprintf("%s:%s", contourIngressContourImage.GetURI(), contourIngressContourImage.Tag),
	)

	features.Register(&manifest)
}
