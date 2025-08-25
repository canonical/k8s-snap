package coredns

import (
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		CoreDNSImage().String(),
	)
}
