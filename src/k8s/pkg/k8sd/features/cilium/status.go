package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckNetwork(ctx context.Context, snap snap.Snap) (bool, error) {
	client, err := snap.KubernetesClient("kube-system")
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ciliumPods := map[string]string{
		"cilium-operator": "io.cilium/app=operator",
		"cilium":          "k8s-app=cilium",
	}

	for ciliumPod, selector := range ciliumPods {
		isReady, err := client.IsPodReady(ctx, ciliumPod, "kube-system", metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, fmt.Errorf("failed to check if pod %q is ready: %w", ciliumPod, err)
		}
		if !isReady {
			return false, fmt.Errorf("cilium pod %q is not yet ready: %w", ciliumPod, err)
		}
	}

	return true, nil
}
