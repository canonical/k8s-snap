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

func TestGetServiceClusterIP(t *testing.T) {
	tests := []struct {
		name           string
		serviceObjects []runtime.Object
		serviceName    string
		namespace      string
		expectedIP     string
		expectError    bool
	}{
		{
			name: "service exists",
			serviceObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP: "192.168.1.1",
					},
				},
			},
			serviceName: "test-service",
			namespace:   "default",
			expectedIP:  "192.168.1.1",
			expectError: false,
		},
		{
			name:           "service does not exist",
			serviceObjects: []runtime.Object{},
			serviceName:    "nonexistent-service",
			namespace:      "default",
			expectedIP:     "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			clientset := fake.NewSimpleClientset(tt.serviceObjects...)
			client := &Client{Interface: clientset}

			ip, err := client.GetServiceClusterIP(context.Background(), tt.serviceName, tt.namespace)

			if tt.expectError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(ip).To(BeEmpty())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ip).To(Equal(tt.expectedIP))
			}
		})
	}
}
