package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// DrainNode drains the specified node by evicting its pods gracefully.
func (c *Client) DrainNode(ctx context.Context, node string) error {
	pods, err := c.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		return fmt.Errorf("failed to get pods for node %s: %w", node, err)
	}
	for _, pod := range pods.Items {
		if pod.Namespace == "kube-system" {
			continue
		} else {
			if err := c.EvictPod(ctx, pod.Namespace, pod.Name); err != nil {
				return fmt.Errorf("failed to evict pod %s from namespace %s: %w", pod.Name, pod.Namespace, err)
			}
		}
	}
	return nil
}

// CordonNode will mark a node as unshedulable.
func (c *Client) CordonNode(ctx context.Context, name string) error {
	node, err := c.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", name, err)
	}

	node.Spec.Unschedulable = true

	if _, err := c.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update node %s: %w", name, err)
	}
	return nil
}

// UncordonNode will mark a node as shedulable.
func (c *Client) UncordonNode(ctx context.Context, name string) error {
	node, err := c.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", name, err)
	}

	node.Spec.Unschedulable = false

	if _, err := c.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update node %s: %w", name, err)
	}
	return nil
}

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

// GracefullyDeleteNode will remove a node from the kubernetes cluster.
// GracefullyDeleteNode will first drain the node to make sure no workloads are running it.
func (c *Client) GracefullyDeleteNode(ctx context.Context, nodeName string) error {
	if err := c.CordonNode(ctx, nodeName); err != nil {
		return fmt.Errorf("failed to cordon node %s: %w", nodeName, err)
	}
	if err := c.DrainNode(ctx, nodeName); err != nil {
		return fmt.Errorf("failed to drain node %s: %w", nodeName, err)
	}
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete node %s: %w", nodeName, err)
		}
		return nil
	})
}
