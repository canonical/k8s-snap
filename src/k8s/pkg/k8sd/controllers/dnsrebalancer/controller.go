package dnsrebalancer

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Controller struct {
	logger           logr.Logger
	client           client.Client
	getClusterConfig func(context.Context) (types.ClusterConfig, error)
	snap             snap.Snap
}

func NewController(
	logger logr.Logger,
	client client.Client,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	snap snap.Snap,
) *Controller {
	return &Controller{
		logger:           logger,
		client:           client,
		getClusterConfig: getClusterConfig,
		snap:             snap,
	}
}

// SetupWithManager sets up the controller with the manager.
func (r *Controller) SetupWithManager(mgr ctrl.Manager) error {
	// Watch Node resources, trigger reconciliation when node status changes
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			// Only trigger reconciliation when a node becomes Ready
			node, ok := object.(*corev1.Node)
			if !ok {
				return false
			}
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		})).
		Complete(r)
}
