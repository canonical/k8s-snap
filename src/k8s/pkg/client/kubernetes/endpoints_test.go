package kubernetes

import (
	"context"
	"errors"
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

func TestGetKubeAPIServerEndpoints(t *testing.T) {
	httpsPortName := "https"
	tests := []struct {
		name              string
		objects           []runtime.Object
		expectedAddresses []string
		expectError       bool
	}{
		{
			name:        "empty",
			expectError: true,
		},
		{
			name: "no endpoints",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints:   []discoveryv1.Endpoint{},
				},
			},
			expectError: true,
		},
		{
			name: "one",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"1.1.1.1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443"},
		},
		{
			name: "two",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"1.1.1.1"}},
						{Addresses: []string{"2.2.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443"},
		},
		{
			name: "multiple-slices",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"1.1.1.1"}},
						{Addresses: []string{"2.2.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes-2",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"3.3.3.3"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:6443"},
		},
		{
			name: "override port",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"1.1.1.1"}},
						{Addresses: []string{"2.2.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes-2",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"3.3.3.3"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(10000))},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:10000"},
		},
		{
			name: "sort",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"3.3.3.3"}},
						{Addresses: []string{"1.1.1.1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(6443))},
					},
				},
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubernetes-2",
						Namespace: "default",
						Labels:    map[string]string{"kubernetes.io/service-name": "kubernetes"},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2.2.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: &httpsPortName, Port: utils.Pointer(int32(10000))},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:10000", "3.3.3.3:6443"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			clientset := fake.NewSimpleClientset(tc.objects...)
			client := &Client{Interface: clientset}

			servers, err := client.GetKubeAPIServerEndpoints(context.Background())
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(servers).To(BeEmpty())
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(servers).To(Equal(tc.expectedAddresses))
			}
		})
	}
}

func TestHasReadyEndpoints(t *testing.T) {
	tests := []struct {
		name          string
		objects       []runtime.Object
		namespace     string
		serviceName   string
		cancelContext bool
		expectedReady bool
		expectError   bool
	}{
		{
			name:          "no endpoints",
			objects:       []runtime.Object{},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: false,
		},
		{
			name: "endpointslice with ready endpoint",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc-abc",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "my-svc",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(true),
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: true,
		},
		{
			name: "endpointslice with unready endpoint",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc-abc",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "my-svc",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(false),
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: false,
		},
		{
			name: "endpointslice with nil ready condition and addresses",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc-abc",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "my-svc",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: true,
		},
		{
			name: "endpointslice with ready condition but empty addresses",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc-abc",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "my-svc",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(true),
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: false,
		},
		{
			name: "corev1 endpoints with ready address",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{IP: "10.0.0.1"},
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: true,
		},
		{
			name: "corev1 endpoints with only not ready address",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							NotReadyAddresses: []corev1.EndpointAddress{
								{IP: "10.0.0.1"},
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: false,
		},
		{
			name: "unready endpointslice does not fall back to ready corev1 endpoints",
			objects: []runtime.Object{
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc-abc",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "my-svc",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(false),
							},
						},
					},
				},
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-svc",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{IP: "10.0.0.1"},
							},
						},
					},
				},
			},
			namespace:     "default",
			serviceName:   "my-svc",
			expectedReady: false,
		},
		{
			name:          "cancelled context returns error",
			objects:       []runtime.Object{},
			namespace:     "default",
			serviceName:   "my-svc",
			cancelContext: true,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			clientset := fake.NewSimpleClientset(tc.objects...)
			client := &Client{Interface: clientset}

			ctx := context.Background()
			if tc.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			ready, err := client.HasReadyEndpoints(ctx, tc.namespace, tc.serviceName)
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(ready).To(Equal(tc.expectedReady))
			}
		})
	}

	t.Run("endpointslice list API error is propagated", func(t *testing.T) {
		g := NewWithT(t)
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("list", "endpointslices", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("forbidden")
		})
		client := &Client{Interface: clientset}

		ready, err := client.HasReadyEndpoints(context.Background(), "default", "my-svc")
		g.Expect(err).To(MatchError(ContainSubstring("failed to list endpoint slices")))
		g.Expect(ready).To(BeFalse())
	})

	t.Run("endpoints get API error is propagated", func(t *testing.T) {
		g := NewWithT(t)
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("get", "endpoints", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("forbidden")
		})
		client := &Client{Interface: clientset}

		ready, err := client.HasReadyEndpoints(context.Background(), "default", "my-svc")
		g.Expect(err).To(MatchError(ContainSubstring("failed to get endpoints")))
		g.Expect(ready).To(BeFalse())
	})
}
