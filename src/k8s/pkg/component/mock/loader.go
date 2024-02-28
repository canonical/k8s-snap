package mock

import "helm.sh/helm/v3/pkg/chart"

type ChartLoader struct {
	Chart *chart.Chart
	Err   error
}

func (m *ChartLoader) Load(path string) (*chart.Chart, error) {
	return m.Chart, m.Err
}
