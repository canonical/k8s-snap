package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckNetwork(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("kube-system")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	for _, check := range []struct {
		name      string
		namespace string
		labels    map[string]string
	}{
		{name: "cilium-operator", namespace: "kube-system", labels: map[string]string{"io.cilium/app": "operator"}},
		{name: "cilium", namespace: "kube-system", labels: map[string]string{"k8s-app": "cilium"}},
	} {
		if err := client.CheckForReadyPods(ctx, check.namespace, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: check.labels}),
		}); err != nil {
			return fmt.Errorf("check %v failed: %w", check.name, err)
		}
	}

	return nil
}
