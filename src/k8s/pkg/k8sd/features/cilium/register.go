package cilium

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", ciliumAgentImageRepo, ciliumAgentImageTag),
		fmt.Sprintf("%s:%s", ciliumOperatorImageRepository, ciliumOperatorImageTag),
	)
}
