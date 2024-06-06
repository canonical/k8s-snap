package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckNetwork checks the status of the Calico pods in the Kubernetes cluster.
// It verifies if all the Calico pods in the "tigera-operator" namespace are ready.
// If any pod is not ready, it returns false. Otherwise, it returns true.
func CheckNetwork(ctx context.Context, snap snap.Snap) (bool, error) {
	client, err := snap.KubernetesClient("calico-system")
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	operatorReady, err := client.IsPodReady(ctx, "kube-system", "tigera-operator", metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get calico pods: %w", err)
	}
	if !operatorReady {
		return false, nil
	}

	calicoPods, err := client.ListPods(ctx, "calico-system", metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get calico pods: %w", err)
	}
	calicoApiserverPods, err := client.ListPods(ctx, "calico-apiserver", metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get calico-apiserver pods: %w", err)
	}

	for _, pod := range append(calicoPods, calicoApiserverPods...) {
		isReady, err := client.IsPodReady(ctx, pod.Name, "calico-system", metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to check if pod %q is ready: %w", pod.Name, err)
		}
		if !isReady {
			return false, nil
		}
	}

	return true, nil
}
