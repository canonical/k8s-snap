package app

import (
	"context"
	"fmt"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	microclusterutil "github.com/canonical/k8s/pkg/utils/microcluster"
	microclusterapi "github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/v2/microcluster"
	"github.com/canonical/microcluster/v2/state"
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

func (a *App) NotifyOrForwardFeatureReconcilation(ctx context.Context, s state.State, network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) error {
	isLeader, err := microclusterutil.IsLeader(s)
	if err != nil {
		return fmt.Errorf("failed to check if node is leader: %w", err)
	}

	// If the node is not the leader, we need to forward the reconcilation request to the leader.
	if !isLeader {
		leaderClient, err := s.Leader()
		if err != nil {
			return fmt.Errorf("failed to get leader client: %w", err)
		}

		in := &apiv1.ReconcileFeaturesRequest{
			Network:       network,
			Gateway:       gateway,
			Ingress:       ingress,
			LoadBalancer:  loadBalancer,
			LocalStorage:  localStorage,
			MetricsServer: metricsServer,
			DNS:           dns,
		}

		result := &apiv1.ReconcileFeaturesResponse{}

		if err := leaderClient.Query(ctx, "POST", apiv1.K8sdAPIVersion, microclusterapi.NewURL().Path(strings.Split(apiv1.ReconcileFeaturesRPC, "/")...), in, result); err != nil {
			return fmt.Errorf("failed to trigger feature reconcilation: %w", err)
		}

		return nil
	}

	a.NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns)
	return nil
}

// Ensure App implements api.Provider.
var _ api.Provider = &App{}
