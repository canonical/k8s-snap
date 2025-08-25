package metrics_server

import (
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		MetricsServerImage().String(),
	)
}
