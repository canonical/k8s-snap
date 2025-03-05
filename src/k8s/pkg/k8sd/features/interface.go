package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

type BaseReconciler struct {
	snap       snap.Snap
	helmClient helm.Client
	state      state.State

	notifyUpdateNodeConfigController func()
}

func NewReconciler(snap snap.Snap, helmClient helm.Client, state state.State, notifyUpdateNodeConfigController func()) BaseReconciler {
	return BaseReconciler{
		snap:       snap,
		helmClient: helmClient,
		state:      state,

		notifyUpdateNodeConfigController: notifyUpdateNodeConfigController,
	}
}

func (r *BaseReconciler) Snap() snap.Snap {
	return r.snap
}

func (r *BaseReconciler) HelmClient() helm.Client {
	return r.helmClient
}

func (r *BaseReconciler) State() state.State {
	return r.state
}

func (r *BaseReconciler) NotifyUpdateNodeConfigController() error {
	if r.notifyUpdateNodeConfigController == nil {
		return fmt.Errorf("notifyUpdateNodeConfigController is not set")
	}
	r.notifyUpdateNodeConfigController()
	return nil
}

type Interface interface {
	NewDNSReconciler(BaseReconciler) Reconciler
	NewNetworkReconciler(BaseReconciler) Reconciler
	NewLoadBalancerReconciler(BaseReconciler) Reconciler
	NewIngressReconciler(BaseReconciler) Reconciler
	NewGatewayReconciler(BaseReconciler) Reconciler
	NewMetricsServerReconciler(BaseReconciler) Reconciler
	NewLocalStorageReconciler(BaseReconciler) Reconciler
}

type Reconciler interface {
	Reconcile(context.Context, types.ClusterConfig) (types.FeatureStatus, error)
}
