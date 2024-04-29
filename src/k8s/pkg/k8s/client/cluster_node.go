package client

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sdClient) JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error {
	if err := c.m.Ready(ctx); err != nil {
		return fmt.Errorf("k8sd API is not ready: %w", err)
	}

	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "join"), request, nil); err != nil {
		// TODO(neoaggelos): only return error that join cluster failed
		return fmt.Errorf("failed to POST /k8sd/cluster/join: %w", err)
	}

	if err := c.WaitForMicroclusterNodeToBeReady(ctx, request.Name); err != nil {
		return fmt.Errorf("microcluster node did not become ready: %w", err)
	}
	return nil
}

func (c *k8sdClient) DeleteClusterMember(ctx context.Context, request apiv1.RemoveNodeRequest) error {
	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "remove"), request, nil); err != nil {
		return fmt.Errorf("failed to POST /k8sd/cluster/remove: %w", err)
	}
	return nil
}

func (c *k8sdClient) CleanupKubernetesServices(ctx context.Context) error {
	if err := c.mc.Query(ctx, "DELETE", api.NewURL().Path("k8sd", "cluster"), nil, nil); err != nil {
		return fmt.Errorf("failed to DELETE /k8sd/cluster: %w", err)
	}
	return nil
}

func (c *k8sdClient) ResetClusterMember(ctx context.Context, name string, force bool) error {
	if err := c.mc.ResetClusterMember(ctx, name, force); err != nil {
		return fmt.Errorf("failed to ResetClusterMember: %w", err)
	}
	return nil
}

// WaitForMicroclusterNodeToBeReady waits until the underlying dqlite node of the microcluster is not in PENDING state.
// While microcluster checkReady will validate that the nodes API server is ready, it will not check if the
// dqlite node is properly setup yet.
func (c *k8sdClient) WaitForMicroclusterNodeToBeReady(ctx context.Context, nodeName string) error {
	return control.WaitUntilReady(ctx, func() (bool, error) {
		clusterStatus, err := c.ClusterStatus(ctx, false)
		if err != nil {
			return false, fmt.Errorf("failed to get the cluster status: %w", err)
		}

		for _, member := range clusterStatus.Members {
			if member.Name == nodeName {
				if member.DatastoreRole == apiv1.DatastoreRolePending {
					return false, nil
				}
				return true, nil
			}
		}
		return false, fmt.Errorf("cluster does not contain node %s", nodeName)
	})
}
