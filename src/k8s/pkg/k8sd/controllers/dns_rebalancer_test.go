package controllers

import (
	"context"
	"testing"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

// getFakeClient returns our wrapper-compatible client from a fake clientset.
func getFakeClient(objects ...runtime.Object) *kubernetes.Client {
	cs := fake.NewSimpleClientset(objects...)
	return &kubernetes.Client{
		Interface: cs,
	}
}

func TestCoreDNSNeedsRebalancing_AllPodsSameNode(t *testing.T) {
	ctx := context.Background()

	pods := []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-1", Labels: map[string]string{"k8s-app": "coredns"}}, Spec: corev1.PodSpec{NodeName: "node-a"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-2", Labels: map[string]string{"k8s-app": "coredns"}}, Spec: corev1.PodSpec{NodeName: "node-a"}},
	}

	fakeClient := getFakeClient(pods...)

	c := &DNSRebalancerController{}
	need, err := c.coreDNSNeedsRebalancing(ctx, fakeClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !need {
		t.Fatalf("expected rebalancing needed when all pods on same node, got %v", need)
	}
}

func TestCoreDNSNeedsRebalancing_Distributed(t *testing.T) {
	ctx := context.Background()

	pods := []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-1", Labels: map[string]string{"k8s-app": "coredns"}}, Spec: corev1.PodSpec{NodeName: "node-a"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-2", Labels: map[string]string{"k8s-app": "coredns"}}, Spec: corev1.PodSpec{NodeName: "node-b"}},
	}

	fakeClient := getFakeClient(pods...)

	c := &DNSRebalancerController{}
	need, err := c.coreDNSNeedsRebalancing(ctx, fakeClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if need {
		t.Fatalf("expected no rebalancing when pods distributed across nodes, got %v", need)
	}
}

func TestCoreDNSNeedsRebalancing_IgnoresUnscheduledPods(t *testing.T) {
	ctx := context.Background()

	// One scheduled, one pending (no NodeName). Should not report imbalance.
	pods := []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-1", Labels: map[string]string{"k8s-app": "coredns"}}, Spec: corev1.PodSpec{NodeName: "node-a"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-2", Labels: map[string]string{"k8s-app": "coredns"}}},
	}

	fakeClient := getFakeClient(pods...)

	c := &DNSRebalancerController{}
	need, err := c.coreDNSNeedsRebalancing(ctx, fakeClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if need {
		t.Fatalf("expected no rebalancing when not all pods are scheduled, got %v", need)
	}

	// Both pending (no NodeName). Should not report imbalance.
	pods2 := []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-3", Labels: map[string]string{"k8s-app": "coredns"}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns-4", Labels: map[string]string{"k8s-app": "coredns"}}},
	}
	fakeClient2 := getFakeClient(pods2...)
	need2, err := c.coreDNSNeedsRebalancing(ctx, fakeClient2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if need2 {
		t.Fatalf("expected no rebalancing when fewer than two pods are scheduled, got %v", need2)
	}
}
