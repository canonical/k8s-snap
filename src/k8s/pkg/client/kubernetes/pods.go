package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitForPodRunning waits for a pod to be in the Running state.
func (c *Client) WaitForPodRunning(ctx context.Context, namespace string, listOptions metav1.ListOptions) error {
	for {
		watcher, err := c.CoreV1().Pods(namespace).Watch(ctx, listOptions)
		if err != nil {
			return fmt.Errorf("failed to watch pod: %w", err)
		}

		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
		}
	}
}
