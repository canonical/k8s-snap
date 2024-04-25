package app

import (
	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/microcluster"
)

func (a *App) MicroCluster() *microcluster.MicroCluster {
	return a.microCluster
}

func (a *App) Snap() snap.Snap {
	return a.snap
}

func (a *App) NotifyUpdateNodeConfigController() {
	utils.MaybeNotify(a.triggerUpdateNodeConfigControllerCh)
}

func (a *App) NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) {
	if network || gateway || ingress || loadBalancer {
		utils.MaybeNotify(a.triggerFeatureControllerNetworkCh)
	}
	if localStorage {
		utils.MaybeNotify(a.triggerFeatureControllerLocalStorageCh)
	}
	if metricsServer {
		utils.MaybeNotify(a.triggerFeatureControllerMetricsServerCh)
	}
	if dns {
		utils.MaybeNotify(a.triggerFeatureControllerDNSCh)
	}
}

// Ensure App implements api.Provider
var _ api.Provider = &App{}
