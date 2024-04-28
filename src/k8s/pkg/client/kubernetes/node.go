package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// DeleteNode will remove a node from the kubernetes cluster.
// DeleteNode will retry if there is a conflict on the resource.
func (c *Client) DeleteNode(ctx context.Context, nodeName string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete node: %w", err)
		}
		return nil
	})
}
