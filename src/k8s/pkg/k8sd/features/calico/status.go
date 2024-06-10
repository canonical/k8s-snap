package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podIsReady(pod v1.Pod) bool {
	if pod.Status.Phase != v1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}

// CheckNetwork checks the status of the Calico pods in the Kubernetes cluster.
// We verify that the tigera-operator and calico-node pods are Ready and in Running state.
func CheckNetwork(ctx context.Context, snap snap.Snap) (bool, error) {
	client, err := snap.KubernetesClient("calico-system")
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
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
		pods, err := client.ListPods(ctx, check.namespace, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: check.labels}),
		})
		if err != nil {
			return false, fmt.Errorf("failed to get %v pods: %w", check.name, err)
		}
		if len(pods) == 0 {
			return false, fmt.Errorf("no %v pods exist on the cluster", check.name)
		}

		for _, pod := range pods {
			if !podIsReady(pod) {
				return false, fmt.Errorf("%v pod %q not ready", check.name, pod.Name)
			}
		}
	}

	return true, nil
}
