package coredns

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckDNS checks the CoreDNS deployment in the cluster.
func CheckDNS(ctx context.Context, snap snap.Snap) (bool, error) {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	isReady, err := client.IsPodReady(ctx, "coredns", "kube-system", metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=coredns"})
	if err != nil {
		return false, fmt.Errorf("failed to wait for CoreDNS pod to be ready: %w", err)
	}

	return isReady, nil
}
