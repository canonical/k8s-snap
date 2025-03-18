package contour

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", ContourIngressEnvoyImageRepo, ContourIngressEnvoyImageTag),
		fmt.Sprintf("%s:%s", ContourIngressContourImageRepo, ContourIngressContourImageTag),
		fmt.Sprintf("%s:%s", ContourGatewayProvisionerContourImageRepo, ContourGatewayProvisionerContourImageTag),
		fmt.Sprintf("%s:%s", ContourGatewayProvisionerEnvoyImageRepo, ContourGatewayProvisionerEnvoyImageTag),
	)
}
