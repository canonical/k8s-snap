package kubernetes

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// DeleteNode will remove a node from the kubernetes cluster.
// DeleteNode will retry if there is a conflict on the resource.
// DeleteNode will not fail if the node does not exist.
func (c *Client) DeleteNode(ctx context.Context, nodeName string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := c.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete node: %w", err)
		}
		return nil
	})
}

func (c *Client) WatchNode(ctx context.Context, name string, reconcile func(node *v1.Node) error) error {
	log := log.FromContext(ctx).WithValues("name", name)
	w, err := c.CoreV1().Nodes().Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return fmt.Errorf("failed to watch node name=%s: %w", name, err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-w.ResultChan():
			if !ok {
				return fmt.Errorf("watch closed")
			}
			node, ok := evt.Object.(*v1.Node)
			if !ok {
				return fmt.Errorf("expected a Node but received %#v", evt.Object)
			}

			if err := reconcile(node); err != nil {
				log.Error(err, "Reconcile Node failed")
			}
		}
	}
}
