package app

import (
	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/microcluster"
)

func (a *App) MicroCluster() *microcluster.MicroCluster {
	return a.microCluster
}

func (a *App) Snap() snap.Snap {
	return a.snap
}

func (a *App) NotifyNodeConfigController() {
	select {
	case a.triggerUpdateNodeConfigControllerCh <- struct{}{}:
	default:
	}
}

// Ensure App implements api.Provider
var _ api.Provider = &App{}
