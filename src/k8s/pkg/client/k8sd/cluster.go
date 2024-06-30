package k8sd

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) BootstrapCluster(ctx context.Context, request apiv1.PostClusterBootstrapRequest) (apiv1.NodeStatus, error) {
	if err := c.app.Ready(ctx); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("k8sd is not ready: %w", err)
	}

	// NOTE(neoaggelos): microcluster adds an arbitrary 30 second timeout in case no context deadline is set.
	// Configure a client deadline for timeout + 30 seconds (the timeout will come from the server)
	ctx, cancel := context.WithTimeout(ctx, request.Timeout+30*time.Second)
	defer cancel()

	var response apiv1.NodeStatus
	if err := c.client.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster"), request, &response); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to POST /k8sd/cluster: %w", err)
	}

	return response, nil
}

func (c *k8sd) JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error {
	if err := c.app.Ready(ctx); err != nil {
		return fmt.Errorf("k8sd is not ready: %w", err)
	}

	// NOTE(neoaggelos): microcluster adds an arbitrary 30 second timeout in case no context deadline is set.
	// Configure a client deadline for timeout + 30 seconds (the timeout will come from the server)
	ctx, cancel := context.WithTimeout(ctx, request.Timeout+30*time.Second)
	defer cancel()

	if err := c.client.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "join"), request, nil); err != nil {
		return fmt.Errorf("failed to POST /k8sd/cluster/join: %w", err)
	}

	// NOTE(neoaggelos): we should not ignore this error
	_ = control.WaitUntilReady(ctx, func() (bool, error) {
		nodeStatus, err := c.NodeStatus(ctx)
		switch {
		case err != nil:
			return false, fmt.Errorf("failed to get node status: %w", err)
		case nodeStatus.DatastoreRole == apiv1.DatastoreRolePending:
			// still waiting for node to join
			return false, nil
		}
		return true, nil
	})

	return nil
}

func (c *k8sd) RemoveNode(ctx context.Context, request apiv1.RemoveNodeRequest) error {
	if err := c.client.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "remove"), request, nil); err != nil {
		return fmt.Errorf("failed to POST /k8sd/cluster/remove: %w", err)
	}
	return nil
}

func (c *k8sd) GetJoinToken(ctx context.Context, request apiv1.GetJoinTokenRequest) (apiv1.GetJoinTokenResponse, error) {
	var response apiv1.GetJoinTokenResponse
	if err := c.client.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster", "tokens"), request, &response); err != nil {
		return apiv1.GetJoinTokenResponse{}, fmt.Errorf("failed to POST /k8sd/cluster/tokens: %w", err)
	}
	return response, nil
}
