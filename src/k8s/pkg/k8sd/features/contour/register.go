package contour

import (
	"github.com/canonical/k8s/pkg/k8sd/charts"
)

func init() {
	charts.Register(&ChartFS)
}
