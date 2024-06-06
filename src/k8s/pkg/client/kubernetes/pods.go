package kubernetes

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsPodReady checks if a pod is ready.
func (c *Client) IsPodReady(ctx context.Context, name, namespace string, listOptions metav1.ListOptions) (bool, error) {
	pods, err := c.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return false, fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, name) {
			if pod.Status.Phase != corev1.PodRunning {
				return false, nil
			}

			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// ListPods lists all pods in a namespace.
func (c *Client) ListPods(ctx context.Context, namespace string, listOptions metav1.ListOptions) ([]corev1.Pod, error) {
	pods, err := c.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	return pods.Items, nil
}
