package api

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/microcluster"
	"github.com/canonical/microcluster/v2/state"
)

// Provider is an interface for state that the API endpoints need access to.
type Provider interface {
	MicroCluster() *microcluster.MicroCluster
	Snap() snap.Snap
	NotifyUpdateNodeConfigController()
	NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool)
	NotifyOrForwardFeatureReconcilation(ctx context.Context, s state.State, network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) error
}
