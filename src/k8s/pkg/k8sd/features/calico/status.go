package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckNetwork checks the status of the Calico pods in the Kubernetes cluster.
// We verify that the tigera-operator and calico-node pods are Ready and in Running state.
func CheckNetwork(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("calico-system")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	for _, check := range []struct {
		name      string
		namespace string
		labels    map[string]string
	}{
		// check that the tigera-operator pods are ready
		{name: "tigera-operator", namespace: "tigera-operator", labels: map[string]string{"k8s-app": "tigera-operator"}},
		// check that calico-node pods are ready
		{name: "calico-node", namespace: "calico-system", labels: map[string]string{"app.kubernetes.io/name": "calico-node"}},
	} {
		if err := client.CheckForReadyPods(ctx, check.namespace, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: check.labels}),
		}); err != nil {
			return fmt.Errorf("check %v failed: %w", check.name, err)
		}
	}

	return nil
}
