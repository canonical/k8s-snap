package k8sd

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

func (c *k8sd) BootstrapCluster(ctx context.Context, request apiv1.BootstrapClusterRequest) (apiv1.BootstrapClusterResponse, error) {
	if err := c.app.Ready(ctx); err != nil {
		return apiv1.BootstrapClusterResponse{}, fmt.Errorf("k8sd is not ready: %w", err)
	}

	// NOTE(neoaggelos): microcluster adds an arbitrary 30 second timeout in case no context deadline is set.
	// Configure a client deadline for timeout + 30 seconds (the timeout will come from the server)
	ctx, cancel := context.WithTimeout(ctx, request.Timeout+30*time.Second)
	defer cancel()

	return query(ctx, c, "POST", apiv1.BootstrapClusterRPC, request, &apiv1.BootstrapClusterResponse{})
}

func (c *k8sd) JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error {
	if err := c.app.Ready(ctx); err != nil {
		return fmt.Errorf("k8sd is not ready: %w", err)
	}

	// NOTE(neoaggelos): microcluster adds an arbitrary 30 second timeout in case no context deadline is set.
	// Configure a client deadline for timeout + 30 seconds (the timeout will come from the server)
	ctx, cancel := context.WithTimeout(ctx, request.Timeout+30*time.Second)
	defer cancel()

	_, err := query(ctx, c, "POST", apiv1.JoinClusterRPC, request, &apiv1.JoinClusterResponse{})
	return err
}

func (c *k8sd) RemoveNode(ctx context.Context, request apiv1.RemoveNodeRequest) error {
	// NOTE(neoaggelos): microcluster adds an arbitrary 30 second timeout in case no context deadline is set.
	// Configure a client deadline for timeout + 30 seconds (the timeout will come from the server)
	ctx, cancel := context.WithTimeout(ctx, request.Timeout+30*time.Second)
	defer cancel()

	_, err := query(ctx, c, "POST", apiv1.RemoveNodeRPC, request, &apiv1.RemoveNodeResponse{})
	return err
}

func (c *k8sd) GetJoinToken(ctx context.Context, request apiv1.GetJoinTokenRequest) (apiv1.GetJoinTokenResponse, error) {
	return query(ctx, c, "POST", apiv1.GetJoinTokenRPC, request, &apiv1.GetJoinTokenResponse{})
}
