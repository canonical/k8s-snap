package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"
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

// NodeVersions returns a map of node names to their parsed Kubernetes versions.
func (c *Client) NodeVersions(ctx context.Context) (map[string]*versionutil.Version, error) {
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeVersions := make(map[string]*versionutil.Version)
	for _, node := range nodes.Items {
		v, err := versionutil.ParseGeneric(node.Status.NodeInfo.KubeletVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version for node %s: %w", node.Name, err)
		}
		nodeVersions[node.Name] = v
	}

	return nodeVersions, nil
}
