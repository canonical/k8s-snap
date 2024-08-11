package k8sd

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

func (c *k8sd) SetClusterAPIAuthToken(ctx context.Context, request apiv1.ClusterAPISetAuthTokenRequest) error {
	_, err := query(ctx, c, "POST", apiv1.ClusterAPISetAuthTokenRPC, request, &apiv1.ClusterAPIGetJoinTokenResponse{})
	return err
}
