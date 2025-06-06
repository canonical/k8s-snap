package app

import (
	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/microcluster"
)

func (a *App) MicroCluster() *microcluster.MicroCluster {
	return a.cluster
}

func (a *App) Snap() snap.Snap {
	return a.snap
}

func (a *App) NotifyUpdateNodeConfigController() {
	utils.MaybeNotify(a.triggerUpdateNodeConfigControllerCh)
}

func (a *App) NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) {
	if network {
		utils.MaybeNotify(a.triggerFeatureControllerNetworkCh)
	}
	if gateway {
		utils.MaybeNotify(a.triggerFeatureControllerGatewayCh)
	}
	if ingress {
		utils.MaybeNotify(a.triggerFeatureControllerIngressCh)
	}
	if loadBalancer {
		utils.MaybeNotify(a.triggerFeatureControllerLoadBalancerCh)
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

// NotifyNetwork notifies the Network feature to reconcile.
func (a *App) NotifyNetwork() {
	utils.MaybeNotify(a.triggerFeatureControllerNetworkCh)
}

// NotifyGateway notifies the Gateway feature to reconcile.
func (a *App) NotifyGateway() {
	utils.MaybeNotify(a.triggerFeatureControllerGatewayCh)
}

// NotifyIngress notifies the Ingress feature to reconcile.
func (a *App) NotifyIngress() {
	utils.MaybeNotify(a.triggerFeatureControllerIngressCh)
}

// NotifyLoadBalancer notifies the Load Balancer feature to reconcile.
func (a *App) NotifyLoadBalancer() {
	utils.MaybeNotify(a.triggerFeatureControllerLoadBalancerCh)
}

// NotifyLocalStorage notifies the Local Storage feature to reconcile.
func (a *App) NotifyLocalStorage() {
	utils.MaybeNotify(a.triggerFeatureControllerLocalStorageCh)
}

// NotifyMetricsServer notifies the Metrics Server feature to reconcile.
func (a *App) NotifyMetricsServer() {
	utils.MaybeNotify(a.triggerFeatureControllerMetricsServerCh)
}

// NotifyDNS notifies the DNS feature to reconcile.
func (a *App) NotifyDNS() {
	utils.MaybeNotify(a.triggerFeatureControllerDNSCh)
}

// Ensure App implements api.Provider.
var _ api.Provider = &App{}
