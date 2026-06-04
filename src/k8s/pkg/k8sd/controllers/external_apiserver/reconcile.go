package external_apiserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconcile maintains the managed Service and EndpointSlices for the "service" backend.
func (r *controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("controller", "external-apiserver")

	config, err := r.getClusterConfig(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get cluster config: %w", err)
	}

	// Backend guard: only the "service" backend manages cluster resources. For any other backend
	// (including the default "external"), ensure the managed resources are absent and stop.
	if config.ControlPlaneEndpoint.GetBackend() != types.ControlPlaneEndpointBackendService {
		if err := r.cleanup(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to clean up managed resources: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Fail closed while no LoadBalancer feature is ready: the Service would otherwise sit Pending
	// and churn. Requeue so we pick up once the feature is enabled.
	if !config.LoadBalancer.GetEnabled() {
		log.V(1).Info("LoadBalancer feature is not enabled, skipping external apiserver reconcile")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Read the authoritative live apiserver set. It is owned by the kube-apiserver and must only
	// be read here, never mutated.
	var endpoints corev1.Endpoints
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: kubernetesEndpointsNamespace, Name: kubernetesEndpointsName}, &endpoints); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get %s/%s endpoints: %w", kubernetesEndpointsNamespace, kubernetesEndpointsName, err)
	}

	host := config.ControlPlaneEndpoint.GetHost()
	externalPort := config.ControlPlaneEndpoint.GetPort()
	backendPort := config.APIServer.GetSecurePort()
	if backendPort == 0 {
		backendPort = 6443
	}

	svc, err := r.reconcileService(ctx, host, externalPort, backendPort)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile service: %w", err)
	}

	ipv4, ipv6 := bucketAddressesByFamily(&endpoints)
	if err := r.reconcileEndpointSlice(ctx, svc, discoveryv1.AddressTypeIPv4, ipv4, backendPort); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile IPv4 endpoint slice: %w", err)
	}
	if err := r.reconcileEndpointSlice(ctx, svc, discoveryv1.AddressTypeIPv6, ipv6, backendPort); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile IPv6 endpoint slice: %w", err)
	}

	return ctrl.Result{}, nil
}

// reconcileService creates or updates the selectorless LoadBalancer Service that fronts the
// control-plane nodes and returns the live object (with UID populated, for owner references).
func (r *controller) reconcileService(ctx context.Context, host string, externalPort, backendPort int) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: serviceNamespace},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, svc, func() error {
		if svc.Labels == nil {
			svc.Labels = map[string]string{}
		}
		svc.Labels[managedByLabel] = managedByValue

		// MetalLB needs an IP to honour the requested VIP. For a DNS host the operator points DNS
		// at the assigned LB IP (or pins it via pool config), so we do not set the annotation.
		if isIP(host) {
			if svc.Annotations == nil {
				svc.Annotations = map[string]string{}
			}
			svc.Annotations[metalLBIPsAnnotation] = host
		} else {
			delete(svc.Annotations, metalLBIPsAnnotation)
		}

		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
		// Selectorless: the apiservers are systemd services, not pods, so kube-proxy cannot derive
		// backends from a selector. The hand-managed EndpointSlices provide them instead.
		svc.Spec.Selector = nil
		// Local preserves the client source IP so apiserver audit logs stay useful (mandated by spec).
		svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyLocal
		svc.Spec.Ports = []corev1.ServicePort{{
			Name:       portName,
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(externalPort),
			TargetPort: intstr.FromInt32(int32(backendPort)),
		}}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// reconcileEndpointSlice creates, updates, or deletes the single-family EndpointSlice for the
// given address type. When addrs is empty the slice is removed.
func (r *controller) reconcileEndpointSlice(ctx context.Context, svc *corev1.Service, family discoveryv1.AddressType, addrs []string, backendPort int) error {
	name := fmt.Sprintf("%s-%s", serviceName, sliceFamilySuffix(family))

	if len(addrs) == 0 {
		return r.deleteIgnoreNotFound(ctx, &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: serviceNamespace},
		})
	}

	slice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: serviceNamespace},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, slice, func() error {
		if slice.Labels == nil {
			slice.Labels = map[string]string{}
		}
		// LabelServiceName binds this slice to the selectorless Service.
		slice.Labels[discoveryv1.LabelServiceName] = serviceName
		slice.Labels[managedByLabel] = managedByValue

		slice.AddressType = family
		slice.Endpoints = make([]discoveryv1.Endpoint, 0, len(addrs))
		for _, addr := range addrs {
			slice.Endpoints = append(slice.Endpoints, discoveryv1.Endpoint{
				Addresses:  []string{addr},
				Conditions: discoveryv1.EndpointConditions{Ready: utils.Pointer(true)},
			})
		}
		// The port name must match the Service port name; kube-proxy uses the slice port number as
		// the backend target.
		slice.Ports = []discoveryv1.EndpointPort{{
			Name:     utils.Pointer(portName),
			Protocol: utils.Pointer(corev1.ProtocolTCP),
			Port:     utils.Pointer(int32(backendPort)),
		}}

		// Own the slice from the Service so it is garbage-collected with it.
		return controllerutil.SetControllerReference(svc, slice, r.client.Scheme())
	})
	return err
}

// cleanup removes all resources managed by this controller. Used when the backend is not
// "service" (idempotent; missing resources are ignored).
func (r *controller) cleanup(ctx context.Context) error {
	objs := []client.Object{
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: serviceNamespace}},
		&discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", serviceName, sliceFamilySuffix(discoveryv1.AddressTypeIPv4)), Namespace: serviceNamespace}},
		&discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", serviceName, sliceFamilySuffix(discoveryv1.AddressTypeIPv6)), Namespace: serviceNamespace}},
	}
	for _, obj := range objs {
		if err := r.deleteIgnoreNotFound(ctx, obj); err != nil {
			return err
		}
	}
	return nil
}

func (r *controller) deleteIgnoreNotFound(ctx context.Context, obj client.Object) error {
	if err := r.client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// bucketAddressesByFamily extracts the ready apiserver IPs from the default/kubernetes Endpoints,
// grouped by address family (an EndpointSlice is single-family).
func bucketAddressesByFamily(endpoints *corev1.Endpoints) (ipv4 []string, ipv6 []string) {
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			if addr.IP == "" {
				continue
			}
			if utils.IsIPv4(addr.IP) {
				ipv4 = append(ipv4, addr.IP)
			} else {
				ipv6 = append(ipv6, addr.IP)
			}
		}
	}
	return ipv4, ipv6
}

func sliceFamilySuffix(family discoveryv1.AddressType) string {
	if family == discoveryv1.AddressTypeIPv6 {
		return "ipv6"
	}
	return "ipv4"
}

func isIP(host string) bool {
	return net.ParseIP(host) != nil
}
