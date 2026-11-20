// Package dnsrebalancer restarts CoreDNS when its replicas are all on one node,
// without repeatedly restarting an unsuccessful rollout.
package dnsrebalancer

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type controller struct {
	logger           logr.Logger
	client           client.Client
	getClusterConfig func(context.Context) (types.ClusterConfig, error)
	now              func() time.Time
	rebalance        *rebalanceAttempt
}

func NewController(
	logger logr.Logger,
	client client.Client,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
) *controller {
	return &controller{
		logger:           logger,
		client:           client,
		getClusterConfig: getClusterConfig,
		now:              time.Now,
	}
}

// SetupWithManager sets up the controller with the manager.
func (r *controller) SetupWithManager(mgr ctrl.Manager) error {
	// All watches reconcile the same cluster-wide CoreDNS placement.
	return ctrl.NewControllerManagedBy(mgr).
		Named("dns-rebalancer").
		Watches(&corev1.Node{}, coreDNSRequestHandler(), builder.WithPredicates(nodeEventPredicate())).
		Watches(&corev1.Pod{}, coreDNSRequestHandler(), builder.WithPredicates(coreDNSPodEventPredicate())).
		Watches(&appsv1.Deployment{}, coreDNSRequestHandler(), builder.WithPredicates(coreDNSDeploymentEventPredicate())).
		Complete(r)
}

func coreDNSRequestHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(context.Context, client.Object) []ctrl.Request {
		return []ctrl.Request{{NamespacedName: k8stypes.NamespacedName{
			Namespace: coredns.Chart.Namespace,
			Name:      coredns.DeploymentName,
		}}}
	})
}

func nodeEventPredicate() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			node, ok := e.Object.(*corev1.Node)
			return ok && nodeSchedulableForCoreDNS(node)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode, ok := e.ObjectOld.(*corev1.Node)
			if !ok {
				return false
			}
			newNode, ok := e.ObjectNew.(*corev1.Node)
			if !ok {
				return false
			}
			return nodeSchedulableForCoreDNS(oldNode) != nodeSchedulableForCoreDNS(newNode)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, ok := e.Object.(*corev1.Node)
			return ok
		},
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

// coreDNSPodEventPredicate watches changes affecting placement or settling.
func coreDNSPodEventPredicate() predicate.Funcs {
	selector := labels.SelectorFromSet(corednsPodLabels)
	isCoreDNSPod := func(obj client.Object) bool {
		pod, ok := obj.(*corev1.Pod)
		return ok && pod.Namespace == coredns.Chart.Namespace && selector.Matches(labels.Set(pod.Labels))
	}
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isCoreDNSPod(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, ok := e.ObjectOld.(*corev1.Pod)
			if !ok {
				return false
			}
			newPod, ok := e.ObjectNew.(*corev1.Pod)
			if !ok {
				return false
			}
			oldMatches, newMatches := isCoreDNSPod(oldPod), isCoreDNSPod(newPod)
			if oldMatches != newMatches {
				return true
			}
			return oldMatches && (oldPod.Spec.NodeName != newPod.Spec.NodeName ||
				kubernetes.IsPodReady(oldPod) != kubernetes.IsPodReady(newPod) ||
				oldPod.Status.Phase != newPod.Status.Phase ||
				(oldPod.DeletionTimestamp == nil) != (newPod.DeletionTimestamp == nil))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isCoreDNSPod(e.Object)
		},
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

func coreDNSDeploymentEventPredicate() predicate.Funcs {
	isCoreDNSDeployment := func(obj client.Object) bool {
		deployment, ok := obj.(*appsv1.Deployment)
		return ok && deployment.Namespace == coredns.Chart.Namespace && deployment.Name == coredns.DeploymentName
	}
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isCoreDNSDeployment(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isCoreDNSDeployment(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isCoreDNSDeployment(e.Object)
		},
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}
