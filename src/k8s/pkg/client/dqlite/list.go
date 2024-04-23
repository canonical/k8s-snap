package dqlite

import (
	"context"
	"fmt"
)

func (c *Client) ListMembers(ctx context.Context) ([]NodeInfo, error) {
	client, err := c.clientGetter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create dqlite client: %w", err)
	}
	defer client.Close()
	return client.Cluster(ctx)
}
