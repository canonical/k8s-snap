package kubernetes_test

import (
	"context"
	"testing"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetNodeName(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		name          string
		nodeName      string
		existingNodes []*corev1.Node // Use a slice of *corev1.Node to represent existing nodes
		expectedError string
	}{
		{
			name:          "node name is empty",
			nodeName:      "",
			existingNodes: []*corev1.Node{},
			expectedError: "node name cannot be empty",
		},
		{
			name:          "node name is available",
			nodeName:      "new-node-name",
			existingNodes: []*corev1.Node{},
			expectedError: "",
		},
		{
			name:     "node name is unavailable",
			nodeName: "existing-node-name",
			existingNodes: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "existing-node-name",
					},
				},
			},
			expectedError: "node name already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a slice of runtime.Object to hold the existing nodes
			var runtimeObjects []runtime.Object
			for _, node := range tc.existingNodes {
				runtimeObjects = append(runtimeObjects, node)
			}

			// Create a fake clientset and add the existing nodes
			clientset := fake.NewSimpleClientset(runtimeObjects...)

			// Create a new k8s client with the fake clientset
			client := &kubernetes.Client{
				Interface: clientset,
			}

			// Call the CheckNodeNameAvailable method
			err := client.CheckNodeNameAvailable(context.Background(), tc.nodeName)
			if tc.expectedError != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tc.expectedError))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}
