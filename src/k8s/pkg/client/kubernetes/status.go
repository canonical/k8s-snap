package kubernetes

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils/control"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitKubernetesEndpointAvailable waits until the kubernetes endpoint becomes available.
func (c *Client) WaitKubernetesEndpointAvailable(ctx context.Context) error {
	return control.WaitUntilReady(ctx, func() (bool, error) {
		// TODO: use the /readyz endpoint instead
		// We want to retry if an error occurs (=API server not ready)
		// returning the error would abort, thus checking for nil
		endpoint, err := c.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{})
		return err == nil && endpoint != nil, nil
	})
}

func (c *Client) CheckKubernetesEndpoint(ctx context.Context) error {
	if endpoint, err := c.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{}); err != nil {
		return fmt.Errorf("failed to get kubernetes endpoint: %w", err)
	} else if endpoint == nil {
		return fmt.Errorf("kubernetes endpoint not found")
	}

	return nil
}

// HasReadyNodes checks the status of all nodes in the Kubernetes cluster.
// HasReadyNodes returns true if there is at least one Ready node in the cluster, false otherwise.
func (c *Client) HasReadyNodes(ctx context.Context) (bool, error) {
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	if err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
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

	return false, nil
}
