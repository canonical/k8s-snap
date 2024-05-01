package kubernetes

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetKubeAPIServerEndpoints(t *testing.T) {
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
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets:    []corev1.EndpointSubset{},
				},
			},
			expectError: true,
		},
		{
			name: "one",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "1.1.1.1"}}},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443"},
		},
		{
			name: "two",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443"},
		},
		{
			name: "multiple-subsets",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
						{Addresses: []corev1.EndpointAddress{{IP: "3.3.3.3"}}},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:6443"},
		},
		{
			name: "override port",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
						{Addresses: []corev1.EndpointAddress{{IP: "3.3.3.3"}}, Ports: []corev1.EndpointPort{{Port: int32(10000), Name: "https"}}},
					},
				},
			},
			expectedAddresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:10000"},
		},
		{
			name: "sort",
			objects: []runtime.Object{
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "3.3.3.3"}, {IP: "1.1.1.1"}}},
						{Addresses: []corev1.EndpointAddress{{IP: "2.2.2.2"}}, Ports: []corev1.EndpointPort{{Port: int32(10000), Name: "https"}}},
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
				g.Expect(err).To(BeNil())
				g.Expect(servers).To(Equal(tc.expectedAddresses))
			}
		})
	}
}
