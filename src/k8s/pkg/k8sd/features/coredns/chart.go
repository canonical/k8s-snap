package coredns

import (
	"embed"
)

//go:embed all:charts
var ChartFS embed.FS
