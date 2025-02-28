package gateway

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	ContourGatewayProvisionerContourImage := FeatureGateway.GetImage(ContourGatewayProvisionerContourImageName)
	ContourGatewayProvisionerEnvoyImage := FeatureGateway.GetImage(ContourGatewayProvisionerEnvoyImageName)

	images.Register(
		fmt.Sprintf("%s:%s", ContourGatewayProvisionerContourImage.GetURI(), ContourGatewayProvisionerContourImage.Tag),
		fmt.Sprintf("%s:%s", ContourGatewayProvisionerEnvoyImage.GetURI(), ContourGatewayProvisionerEnvoyImage.Tag),
	)
}
