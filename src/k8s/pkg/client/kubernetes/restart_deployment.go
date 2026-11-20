package kubernetes

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestartedAtAnnotation triggers a rollout when set on a workload's pod template.
const RestartedAtAnnotation = "kubectl.kubernetes.io/restartedAt"

// SetDeploymentRestartedAt sets the pod template restart timestamp, preserving other annotations.
func SetDeploymentRestartedAt(deployment *appsv1.Deployment, restartedAt time.Time) {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations[RestartedAtAnnotation] = restartedAt.Format(time.RFC3339)
}

// RestartDeployment updates the restartedAt field to trigger a rollout restart for the given Deployment.
func (c *Client) RestartDeployment(ctx context.Context, name, namespace string) error {
	deployment, err := c.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s in namespace %s: %w", name, namespace, err)
	}

	SetDeploymentRestartedAt(deployment, time.Now())

	_, err = c.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to set restartedAt annotation for deployment %s in namespace %s: %w", name, namespace, err)
	}
	return nil
}
