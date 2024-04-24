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

func (a *App) NotifyUpdateNodeConfigController() {
	notify(a.triggerUpdateNodeConfigControllerCh)
}

func (a *App) NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) {
	if network || gateway || ingress || loadBalancer {
		notify(a.triggerFeatureControllerNetworkCh)
	}
	if localStorage {
		notify(a.triggerFeatureControllerLocalStorageCh)
	}
	if metricsServer {
		notify(a.triggerFeatureControllerMetricsServerCh)
	}
	if dns {
		notify(a.triggerFeatureControllerDNSCh)
	}
}

func notify(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// Ensure App implements api.Provider
var _ api.Provider = &App{}
