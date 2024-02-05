package client

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared/api"
)

func (c *Client) CreateJoinToken(ctx context.Context, name string, worker bool) (string, error) {
	if !worker {
		return c.m.NewJoinToken(name)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := apiv1.WorkerNodeTokenRequest{}
	response := apiv1.WorkerNodeTokenResponse{}

	err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "worker", "tokens"), request, &response)
	if err != nil {
		return "", fmt.Errorf("failed to query endpoint POST /k8sd/worker/tokens: %w", err)
	}
	return response.EncodedToken, nil
}
