package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestartDeployment updates the restartedAt field to trigger a rollout restart for the given Deployment.
func RestartDeployment(ctx context.Context, client *k8sClient, name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}
