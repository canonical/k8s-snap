package kubernetes

import (
	"context"
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
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
