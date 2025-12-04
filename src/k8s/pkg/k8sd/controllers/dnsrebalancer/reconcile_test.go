package dnsrebalancer

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcile_LessThanTwoNodesReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// Only one node ready
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(node).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	result, err := reconciler.Reconcile(ctx, ctrl.Request{})

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestReconcile_CoreDNSAlreadyBalanced(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// Two nodes ready
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}

	// Pods already distributed
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-1",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-2",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-2"},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&nodes[0], &nodes[1], &pods[0], &pods[1]).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	result, err := reconciler.Reconcile(ctx, ctrl.Request{})

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCoreDNSNeedsRebalancing_AllPodsSameNode(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-1",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-a"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-2",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-a"},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&pods[0], &pods[1]).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	needsRebalancing, err := reconciler.coreDNSNeedsRebalancing(ctx)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(needsRebalancing).To(BeTrue(), "should need rebalancing when all pods on same node")
}

func TestCoreDNSNeedsRebalancing_Distributed(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-1",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-a"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-2",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-b"},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&pods[0], &pods[1]).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	needsRebalancing, err := reconciler.coreDNSNeedsRebalancing(ctx)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(needsRebalancing).To(BeFalse(), "should not need rebalancing when pods distributed")
}

func TestCoreDNSNeedsRebalancing_IgnoresUnscheduledPods(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// One scheduled, one pending (no NodeName)
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-1",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{NodeName: "node-a"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-2",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&pods[0], &pods[1]).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	needsRebalancing, err := reconciler.coreDNSNeedsRebalancing(ctx)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(needsRebalancing).To(BeFalse(), "should not need rebalancing with unscheduled pods")
}

func TestCoreDNSNeedsRebalancing_AllPodsPending(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// Both pods pending (no NodeName)
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-1",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "coredns-2",
				Labels:    map[string]string{"k8s-app": "coredns"},
			},
			Spec: corev1.PodSpec{},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&pods[0], &pods[1]).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	needsRebalancing, err := reconciler.coreDNSNeedsRebalancing(ctx)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(needsRebalancing).To(BeFalse(), "should not need rebalancing when no pods scheduled")
}

func TestCoreDNSNeedsRebalancing_NoPods(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		Build()

	reconciler := &Controller{
		logger: ctrl.Log.WithName("test"),
		client: fakeClient,
		snap:   nil,
	}

	needsRebalancing, err := reconciler.coreDNSNeedsRebalancing(ctx)

	g.Expect(err).To(HaveOccurred(), "should error when no CoreDNS pods found")
	g.Expect(needsRebalancing).To(BeFalse())
}
