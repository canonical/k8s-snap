package ingress

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	ContourIngressContourImage := FeatureIngress.GetImage(ContourIngressContourImageName)
	ContourIngressEnvoyImage := FeatureIngress.GetImage(ContourIngressEnvoyImageName)

	images.Register(
		fmt.Sprintf("%s:%s", ContourIngressEnvoyImage.GetURI(), ContourIngressEnvoyImage.Tag),
		fmt.Sprintf("%s:%s", ContourIngressContourImage.GetURI(), ContourIngressContourImage.Tag),
	)
}
