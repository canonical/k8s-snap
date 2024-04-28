package kubernetes

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterHasReadyNodes(t *testing.T) {
	tests := []struct {
		name          string
		nodes         []runtime.Object
		expectedReady bool
	}{
		{
			name: "all nodes not ready",
			nodes: []runtime.Object{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionFalse},
						},
					},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionFalse},
						},
					},
				},
			},
			expectedReady: false,
		},
		{
			name: "some nodes not ready",
			nodes: []runtime.Object{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionTrue},
						},
					},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionFalse},
						},
					},
				},
			},
			expectedReady: true,
		},
		{
			name: "all nodes ready",
			nodes: []runtime.Object{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionTrue},
						},
					},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: v1.NodeReady, Status: v1.ConditionTrue},
						},
					},
				},
			},
			expectedReady: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			clientset := fake.NewSimpleClientset(tt.nodes...)
			client := &Client{Interface: clientset}

			ready, err := client.HasReadyNodes(context.Background())

			g.Expect(err).To(BeNil())
			g.Expect(ready).To(Equal(tt.expectedReady))
		})
	}
}
