package metrics_server

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/charts"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	images.Register(
		fmt.Sprintf("%s:%s", imageRepo, imageTag),
	)

	charts.Register(&ChartFS)
}
