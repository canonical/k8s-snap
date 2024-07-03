package contour

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", contourIngressEnvoyImageRepo, contourIngressEnvoyImageTag),
		fmt.Sprintf("%s:%s", contourIngressContourImageRepo, contourIngressContourImageTag),
		fmt.Sprintf("%s:%s", contourGatewayProvisionerContourImageRepo, contourGatewayProvisionerContourImageTag),
		fmt.Sprintf("%s:%s", contourGatewayProvisionerEnvoyImageRepo, contourGatewayProvisionerEnvoyImageTag),
	)
}
