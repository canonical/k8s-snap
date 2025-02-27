package calico

import (
	"embed"
)

//go:embed all:charts
var ChartFS embed.FS
