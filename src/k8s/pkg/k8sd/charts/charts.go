package charts

import (
	"embed"
)

// registeredCharts is a list of filesystems that contain charts for features used by k8s-snap.
var registeredCharts []*embed.FS

// Charts returns the list of registered charts.
func Charts() []*embed.FS {
	if registeredCharts == nil {
		return nil
	}
	chartFSList := make([]*embed.FS, len(registeredCharts))
	copy(chartFSList, registeredCharts)
	return chartFSList
}

// Register charts that are used by k8s-snap.
// Register is used by the `init()` method in individual packages.
func Register(charts *embed.FS) {
	registeredCharts = append(registeredCharts, charts)
}
