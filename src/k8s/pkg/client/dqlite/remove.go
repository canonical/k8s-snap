package dqlite

import (
	"context"
	"fmt"
)

func (c *Client) RemoveNodeByAddress(ctx context.Context, address string) error {
	client, err := c.clientGetter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create dqlite client: %w", err)
	}
	members, err := client.Cluster(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster nodes")
	}

	var (
		memberExists, clusterHasOtherVoters bool
		memberToRemove                      NodeInfo
	)
	for _, member := range members {
		switch {
		case member.Address == address:
			memberToRemove = member
			memberExists = true

		case member.Address != address && member.Role == Voter:
			clusterHasOtherVoters = true
		}
	}

	if !memberExists {
		return fmt.Errorf("cluster does not have a node with address %v", address)
	}

	// TODO: consider using client.Transfer() for a different node to become leader
	if !clusterHasOtherVoters {
		return fmt.Errorf("not removing node because there are no other voter members")
	}

	if err := client.Remove(ctx, memberToRemove.ID); err != nil {
		return fmt.Errorf("failed to remove node %#v from dqlite cluster: %w", memberToRemove, err)
	}

	return nil
}
