package kubernetes

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestartDaemonset updates the restartedAt field to trigger a rollout restart for the given DaemonSet.
func (c *Client) RestartDaemonset(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	daemonset, err := c.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonset %s in namespace %s: %w", name, namespace, err)
	}

	if daemonset.Spec.Template.ObjectMeta.Annotations == nil {
		daemonset.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	daemonset.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = c.AppsV1().DaemonSets(namespace).Update(ctx, daemonset, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to set restartedAt annotation for daemonset %s in namespace %s: %w", name, namespace, err)
	}
	return nil
}
