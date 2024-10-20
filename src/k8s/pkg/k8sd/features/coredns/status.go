package coredns

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckDNS checks the CoreDNS deployment in the cluster.
func CheckDNS(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("kube-system")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	for _, check := range []struct {
		name      string
		namespace string
		labels    map[string]string
	}{
		{name: "coredns", namespace: "kube-system", labels: map[string]string{"app.kubernetes.io/name": "coredns"}},
	} {
		if err := client.CheckForReadyPods(ctx, check.namespace, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: check.labels}),
		}); err != nil {
			return fmt.Errorf("%v pods not yet ready: %w", check.name, err)
		}
	}

	return nil
}
