package metrics_server

import (
	"embed"
)

//go:embed all:charts
var ChartFS embed.FS
