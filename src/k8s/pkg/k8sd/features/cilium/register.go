package cilium

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", CiliumAgentImageRepo, CiliumAgentImageTag),
		fmt.Sprintf("%s-generic:%s", ciliumOperatorImageRepo, ciliumOperatorImageTag),
	)
}
