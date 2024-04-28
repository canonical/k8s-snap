package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceClusterIP retrieves the ClusterIP from a Kubernetes service.
func (c *Client) GetServiceClusterIP(ctx context.Context, name, namespace string) (string, error) {
	svc, err := c.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service '%s' in namespace '%s': %w", name, namespace, err)
	}

	return svc.Spec.ClusterIP, nil
}
