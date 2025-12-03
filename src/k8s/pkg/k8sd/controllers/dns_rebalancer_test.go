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
