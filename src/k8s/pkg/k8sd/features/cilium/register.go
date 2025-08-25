package cilium

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		CiliumAgentImage().String(),
		fmt.Sprintf("%s-generic:%s", CiliumOperatorImage().Repository, CiliumOperatorImage().Tag),
	)
}
