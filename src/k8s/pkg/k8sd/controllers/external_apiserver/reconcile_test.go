package external_apiserver

import (
	"context"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func serviceConfig(host string) types.ClusterConfig {
	return types.ClusterConfig{
		APIServer:    types.APIServer{SecurePort: utils.Pointer(6443)},
		LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(true)},
		ControlPlaneEndpoint: types.ControlPlaneEndpoint{
			Host:    utils.Pointer(host),
			Port:    utils.Pointer(6443),
			Backend: utils.Pointer(types.ControlPlaneEndpointBackendService),
		},
	}
}

func kubernetesEndpoints(ipv4 []string, ipv6 []string) *corev1.Endpoints {
	addrs := make([]corev1.EndpointAddress, 0, len(ipv4)+len(ipv6))
	for _, ip := range ipv4 {
		addrs = append(addrs, corev1.EndpointAddress{IP: ip})
	}
	for _, ip := range ipv6 {
		addrs = append(addrs, corev1.EndpointAddress{IP: ip})
	}
	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: kubernetesEndpointsNamespace, Name: kubernetesEndpointsName},
		Subsets: []corev1.EndpointSubset{{
			Addresses: addrs,
			Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443, Protocol: corev1.ProtocolTCP}},
		}},
	}
}

func newReconciler(c client.Client, cfg types.ClusterConfig) *controller {
	return &controller{
		logger: ctrl.Log.WithName("test"),
		client: c,
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return cfg, nil
		},
	}
}

func getService(g *WithT, c client.Client) *corev1.Service {
	svc := &corev1.Service{}
	err := c.Get(context.Background(), client.ObjectKey{Namespace: serviceNamespace, Name: serviceName}, svc)
	g.Expect(err).ToNot(HaveOccurred())
	return svc
}

func getSlice(g *WithT, c client.Client, suffix string) *discoveryv1.EndpointSlice {
	slice := &discoveryv1.EndpointSlice{}
	err := c.Get(context.Background(), client.ObjectKey{Namespace: serviceNamespace, Name: serviceName + "-" + suffix}, slice)
	g.Expect(err).ToNot(HaveOccurred())
	return slice
}

func sliceAddresses(slice *discoveryv1.EndpointSlice) []string {
	addrs := make([]string, 0, len(slice.Endpoints))
	for _, ep := range slice.Endpoints {
		addrs = append(addrs, ep.Addresses...)
	}
	return addrs
}

func TestReconcile_ServiceBackend_CreatesResources(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(kubernetesEndpoints([]string{"10.0.0.1", "10.0.0.2"}, nil)).
		Build()

	r := newReconciler(fakeClient, serviceConfig("10.0.0.250"))

	result, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))

	svc := getService(g, fakeClient)
	g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
	g.Expect(svc.Spec.Selector).To(BeNil())
	g.Expect(svc.Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyLocal))
	g.Expect(svc.Annotations).To(HaveKeyWithValue(metalLBIPsAnnotation, "10.0.0.250"))
	g.Expect(svc.Spec.Ports).To(HaveLen(1))
	g.Expect(svc.Spec.Ports[0].Name).To(Equal(portName))
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(6443)))
	g.Expect(svc.Spec.Ports[0].TargetPort.IntValue()).To(Equal(6443))

	slice := getSlice(g, fakeClient, "ipv4")
	g.Expect(slice.AddressType).To(Equal(discoveryv1.AddressTypeIPv4))
	g.Expect(slice.Labels).To(HaveKeyWithValue(discoveryv1.LabelServiceName, serviceName))
	g.Expect(sliceAddresses(slice)).To(ConsistOf("10.0.0.1", "10.0.0.2"))
	g.Expect(slice.Ports).To(HaveLen(1))
	g.Expect(*slice.Ports[0].Name).To(Equal(portName))
	g.Expect(*slice.Ports[0].Port).To(Equal(int32(6443)))
	g.Expect(slice.OwnerReferences).To(HaveLen(1))
	g.Expect(slice.OwnerReferences[0].Name).To(Equal(serviceName))

	// no IPv6 slice should exist
	err = fakeClient.Get(ctx, client.ObjectKey{Namespace: serviceNamespace, Name: serviceName + "-ipv6"}, &discoveryv1.EndpointSlice{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
}

func TestReconcile_ServiceBackend_DNSHost_NoAnnotation(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(kubernetesEndpoints([]string{"10.0.0.1"}, nil)).
		Build()

	r := newReconciler(fakeClient, serviceConfig("api.example.com"))

	_, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())

	svc := getService(g, fakeClient)
	g.Expect(svc.Annotations).ToNot(HaveKey(metalLBIPsAnnotation))
}

func TestReconcile_ServiceBackend_DualStack(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(kubernetesEndpoints([]string{"10.0.0.1"}, []string{"fd01::1"})).
		Build()

	r := newReconciler(fakeClient, serviceConfig("10.0.0.250"))

	_, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(sliceAddresses(getSlice(g, fakeClient, "ipv4"))).To(ConsistOf("10.0.0.1"))
	ipv6 := getSlice(g, fakeClient, "ipv6")
	g.Expect(ipv6.AddressType).To(Equal(discoveryv1.AddressTypeIPv6))
	g.Expect(sliceAddresses(ipv6)).To(ConsistOf("fd01::1"))
}

func TestReconcile_ServiceBackend_SyncsOnEndpointChange(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	endpoints := kubernetesEndpoints([]string{"10.0.0.1", "10.0.0.2"}, nil)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(endpoints).
		Build()

	r := newReconciler(fakeClient, serviceConfig("10.0.0.250"))

	_, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(sliceAddresses(getSlice(g, fakeClient, "ipv4"))).To(ConsistOf("10.0.0.1", "10.0.0.2"))

	// A control-plane node leaves: the endpoints object loses an address.
	updated := kubernetesEndpoints([]string{"10.0.0.1"}, nil)
	updated.ResourceVersion = getEndpointsResourceVersion(g, fakeClient)
	g.Expect(fakeClient.Update(ctx, updated)).To(Succeed())

	_, err = r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(sliceAddresses(getSlice(g, fakeClient, "ipv4"))).To(ConsistOf("10.0.0.1"))
}

func getEndpointsResourceVersion(g *WithT, c client.Client) string {
	ep := &corev1.Endpoints{}
	g.Expect(c.Get(context.Background(), client.ObjectKey{Namespace: kubernetesEndpointsNamespace, Name: kubernetesEndpointsName}, ep)).To(Succeed())
	return ep.ResourceVersion
}

func TestReconcile_LoadBalancerDisabled_NoResources(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cfg := serviceConfig("10.0.0.250")
	cfg.LoadBalancer.Enabled = utils.Pointer(false)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(kubernetesEndpoints([]string{"10.0.0.1"}, nil)).
		Build()

	r := newReconciler(fakeClient, cfg)

	result, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.RequeueAfter).To(Equal(30 * time.Second))

	err = fakeClient.Get(ctx, client.ObjectKey{Namespace: serviceNamespace, Name: serviceName}, &corev1.Service{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
}

func TestReconcile_ExternalBackend_CleansUp(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// Pre-existing managed resources (e.g. left over from a previous "service" backend).
	existingSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: serviceNamespace},
		Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer},
	}
	existingSlice := &discoveryv1.EndpointSlice{
		ObjectMeta:  metav1.ObjectMeta{Name: serviceName + "-ipv4", Namespace: serviceNamespace},
		AddressType: discoveryv1.AddressTypeIPv4,
	}

	cfg := serviceConfig("10.0.0.250")
	cfg.ControlPlaneEndpoint.Backend = utils.Pointer(types.ControlPlaneEndpointBackendExternal)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(kubernetesEndpoints([]string{"10.0.0.1"}, nil), existingSvc, existingSlice).
		Build()

	r := newReconciler(fakeClient, cfg)

	result, err := r.Reconcile(ctx, ctrl.Request{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))

	err = fakeClient.Get(ctx, client.ObjectKey{Namespace: serviceNamespace, Name: serviceName}, &corev1.Service{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
	err = fakeClient.Get(ctx, client.ObjectKey{Namespace: serviceNamespace, Name: serviceName + "-ipv4"}, &discoveryv1.EndpointSlice{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
}
