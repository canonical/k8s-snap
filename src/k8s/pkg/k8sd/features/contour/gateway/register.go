package gateway

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	contourGatewayProvisionerContourImage := FeatureGateway.GetImage(ContourGatewayProvisionerContourImageName)
	contourGatewayProvisionerEnvoyImage := FeatureGateway.GetImage(ContourGatewayProvisionerEnvoyImageName)

	images.Register(
		fmt.Sprintf("%s:%s", contourGatewayProvisionerContourImage.GetURI(), contourGatewayProvisionerContourImage.Tag),
		fmt.Sprintf("%s:%s", contourGatewayProvisionerEnvoyImage.GetURI(), contourGatewayProvisionerEnvoyImage.Tag),
	)
}
