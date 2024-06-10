package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckForReadyPods checks if all pods in the specified namespace are ready.
// It returns an error if any of the pods are not ready.
// The listOptions specify additional options for listing the pods, e.g. labels.
// It returns an error if it fails to list the pods or if there are no pods in the namespace.
// If any of the pods are not ready, it returns an error with the names of the not ready pods.
// If all pods are ready, it returns nil.
func (c *Client) CheckForReadyPods(ctx context.Context, namespace string, listOptions metav1.ListOptions) error {
	pods, err := c.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods in %v namespace on the cluster", namespace)
	}

	notReadyPods := []string{}
	for _, pod := range pods.Items {
		if !podIsReady(pod) {
			notReadyPods = append(notReadyPods, pod.Name)
		}
	}

	if len(notReadyPods) > 0 {
		return fmt.Errorf("pods %v not ready", notReadyPods)
	}
	return nil
}

// podIsReady checks if a pod is in the ready state.
// It returns true if the pod is running (Condition "Ready" = true).
// Otherwise, it returns false.
func podIsReady(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
