package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CheckNodeNameAvailable(ctx context.Context, nodeName string) error {
	if nodeName == "" {
		return fmt.Errorf("node name cannot be empty")
	}
	if _, err := c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{}); err == nil {
		return fmt.Errorf("node name already exists: %s", nodeName)
	} else if !apierrors.IsNotFound(err) {
		// Request to fetch node failed for some other reason
		return fmt.Errorf("failed to check whether node name available %s: %w", nodeName, err)
	}
	return nil
}
