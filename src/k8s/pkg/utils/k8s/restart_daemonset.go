package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestartDaemonset updates the restartedAt field to trigger a rollout restart for the given DaemonSet.
func RestartDaemonset(ctx context.Context, client *k8sClient, name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	daemonset, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if daemonset.Spec.Template.ObjectMeta.Annotations == nil {
		daemonset.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	daemonset.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = client.AppsV1().DaemonSets(namespace).Update(ctx, daemonset, metav1.UpdateOptions{})
	return err
}
