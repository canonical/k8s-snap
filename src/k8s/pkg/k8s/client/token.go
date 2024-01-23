package client

import (
	"context"
)

func (c *Client) CreateJoinToken(ctx context.Context, name string) (string, error) {
	return c.m.NewJoinToken(name)
}
