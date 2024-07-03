package contour

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", envoyImageRepo, envoyImageTag),
		fmt.Sprintf("%s:%s", contourImageRepo, contourImageTag),
		fmt.Sprintf("%s:%s", contourGatewayImageRepo, contourGatewayImageTag),
		fmt.Sprintf("%s:%s", envoyProxyImageRepo, envoyProxyImageTag),
	)
}
