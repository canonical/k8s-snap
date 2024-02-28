package mock

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
)

type ChartInstaller struct {
	enabledComponents map[string]bool
}

func (m *ChartInstaller) Run(install *action.Install, chart *chart.Chart, values map[string]any) error {
	m.enabledComponents[install.ReleaseName] = true
	return nil
}

func (m *ChartInstaller) NewInstall(actionConfig *action.Configuration) *action.Install {
	return nil
}
