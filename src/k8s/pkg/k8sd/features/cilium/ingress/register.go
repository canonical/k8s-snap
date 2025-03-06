package ingress

import (
	"github.com/canonical/k8s/pkg/k8sd/features/manifests"
)

func init() {
	manifests.Register(&manifest)
}
