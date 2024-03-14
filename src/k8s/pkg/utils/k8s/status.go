package k8s

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils/control"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) listNodes(ctx context.Context) (*v1.NodeList, error) {
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}
	return nodes, nil
}

func (c *Client) WaitApiServerReady(ctx context.Context) error {
	return control.WaitUntilReady(ctx, func() (bool, error) {
		_, err := c.listNodes(ctx)
		return err == nil, nil
	})
}

func (c *Client) IsClusterReady(ctx context.Context) (bool, error) {
	nodes, err := c.listNodes(ctx)

	if err != nil {
		return false, err
	}

	if len(nodes.Items) == 0 {
		return false, fmt.Errorf("cluster has no nodes")
	}

	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady {
				if condition.Status == v1.ConditionTrue {
					// At least one node is ready.
					return true, nil
				}
			}
		}
	}

	return false, nil // Cluster is not ready but has nodes
}
