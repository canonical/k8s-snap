package ingress

import "github.com/canonical/k8s/pkg/k8sd/features"

func init() {
	features.Register(&manifest)
}
