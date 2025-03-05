package metrics_server

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/charts"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/images"
)

func init() {
	metricsServerImage := Manifest.GetImage(MetricsServerImageName)

	images.Register(
		fmt.Sprintf("%s:%s", metricsServerImage.GetURI(), metricsServerImage.Tag),
	)

	charts.Register(&ChartFS)

	features.Register(&Manifest)
}
