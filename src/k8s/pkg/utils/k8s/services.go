package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceClusterIP retrieves the ClusterIP from a Kubernetes service.
// An empty namespace will default to "default".
func GetServiceClusterIP(ctx context.Context, client *k8sClient, name, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	svc, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service '%s' in namespace '%s': %w", name, namespace, err)
	}

	return svc.Spec.ClusterIP, nil
}
