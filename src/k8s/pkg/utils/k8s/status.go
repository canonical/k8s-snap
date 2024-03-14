package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) IsClusterReady(ctx context.Context) (bool, error) {
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	if err != nil {
		return false, fmt.Errorf("kube-apiserver not ready. failed to get nodes: %v", err)
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

	if len(nodes.Items) == 0 {
		return false, fmt.Errorf("cluster has no nodes")
	}

	return false, nil // Cluster is not ready but has nodes
}
