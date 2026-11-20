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

			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(ready).To(Equal(tt.expectedReady))
		})
	}
}

func TestIsNodeReady(t *testing.T) {
	tests := []struct {
		name string
		node *v1.Node
		want bool
	}{
		{name: "nil"},
		{name: "no conditions", node: &v1.Node{}},
		{name: "unrelated condition", node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{{Type: v1.NodeMemoryPressure, Status: v1.ConditionTrue}}}}},
	}
	for _, status := range []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionFalse, v1.ConditionUnknown} {
		tests = append(tests, struct {
			name string
			node *v1.Node
			want bool
		}{
			name: string(status),
			node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{
				{Type: v1.NodeMemoryPressure, Status: v1.ConditionFalse},
				{Type: v1.NodeReady, Status: status},
			}}},
			want: status == v1.ConditionTrue,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNodeReady(tt.node); got != tt.want {
				t.Fatalf("IsNodeReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
