// Package external_apiserver implements the "service" backend of the control-plane endpoint
// feature. It maintains a selectorless LoadBalancer Service plus hand-managed EndpointSlices in
// kube-system whose backends mirror the live kube-apiserver set from the default/kubernetes
// Endpoints object, requesting the configured VIP from MetalLB.
package external_apiserver

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// serviceName is the name of the managed LoadBalancer Service and the base name of the
	// managed EndpointSlices.
	serviceName = "k8s-external-apiserver"
	// serviceNamespace is the namespace the managed resources live in.
	serviceNamespace = "kube-system"
	// portName is the shared port name on the Service and the EndpointSlices; kube-proxy uses it
	// to bind the Service port to the slice backend port.
	portName = "https"
	// metalLBIPsAnnotation is the MetalLB annotation used to request a specific VIP.
	metalLBIPsAnnotation = "metallb.universe.tf/loadBalancerIPs"
	// managedByLabel marks the resources owned by this controller.
	managedByLabel = "app.kubernetes.io/managed-by"
	managedByValue = "k8sd"

	kubernetesEndpointsName      = "kubernetes"
	kubernetesEndpointsNamespace = "default"
)

type controller struct {
	logger           logr.Logger
	client           client.Client
	getClusterConfig func(context.Context) (types.ClusterConfig, error)
}

// NewController creates a new external apiserver controller.
func NewController(
	logger logr.Logger,
	client client.Client,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
) *controller {
	return &controller{
		logger:           logger,
		client:           client,
		getClusterConfig: getClusterConfig,
	}
}

// SetupWithManager sets up the controller with the manager.
func (r *controller) SetupWithManager(mgr ctrl.Manager) error {
	// Reconcile on changes to the live apiserver set (the default/kubernetes Endpoints object),
	// and self-heal on edits/deletes of our own managed Service/EndpointSlices.
	return ctrl.NewControllerManagedBy(mgr).
		Named("external-apiserver").
		For(&corev1.Endpoints{}, builder.WithPredicates(predicate.NewPredicateFuncs(isKubernetesEndpoints))).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(enqueueKubernetesEndpoints),
			builder.WithPredicates(predicate.NewPredicateFuncs(isManagedObject)),
		).
		Watches(
			&discoveryv1.EndpointSlice{},
			handler.EnqueueRequestsFromMapFunc(enqueueKubernetesEndpoints),
			builder.WithPredicates(predicate.NewPredicateFuncs(isManagedObject)),
		).
		Complete(r)
}

// isKubernetesEndpoints matches only the default/kubernetes Endpoints object.
func isKubernetesEndpoints(o client.Object) bool {
	return o.GetNamespace() == kubernetesEndpointsNamespace && o.GetName() == kubernetesEndpointsName
}

// isManagedObject matches the Service and EndpointSlices managed by this controller.
func isManagedObject(o client.Object) bool {
	if o.GetNamespace() != serviceNamespace {
		return false
	}
	if o.GetName() == serviceName {
		return true
	}
	return o.GetLabels()[discoveryv1.LabelServiceName] == serviceName
}

// enqueueKubernetesEndpoints maps any managed-object event back to a reconcile of the
// default/kubernetes Endpoints, which is the single key this controller reconciles on.
func enqueueKubernetesEndpoints(context.Context, client.Object) []reconcile.Request {
	return []reconcile.Request{{
		NamespacedName: client.ObjectKey{Namespace: kubernetesEndpointsNamespace, Name: kubernetesEndpointsName},
	}}
}
