package dqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/utils/control"
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
		freeSpareNode                       *NodeInfo
	)
	for _, member := range members {
		switch {
		case member.Address == address:
			memberToRemove = member
			memberExists = true

		case member.Address != address && member.Role == Voter:
			clusterHasOtherVoters = true

		case member.Address != address && member.Role == Spare:
			// This is only used in a two node setup, where the leader node is removed.
			// The free spare node will be promoted to leader before the existing leader is removed.
			freeSpareNode = &member
		}
	}

	if !memberExists {
		return fmt.Errorf("cluster does not have a node with address %v", address)
	}

	if !clusterHasOtherVoters {
		if freeSpareNode == nil {
			// This normally should not happen. There should always be a backup node, except
			// if one tries to remove the last node in the cluster.
			return fmt.Errorf("cannot transfer dqlite leadership as there is no remaining spare node")
		}

		// Leadership can only be transfered to a voter or standby node.
		// Therefore the remaining node in the cluster needs to be promoted first.
		if err := client.Assign(ctx, freeSpareNode.ID, Voter); err != nil {
			return fmt.Errorf("failed to assign voter role to %d: %w", freeSpareNode.ID, err)
		}
		// Transfer leadership to remaining node in cluster.
		if err := client.Transfer(ctx, freeSpareNode.ID); err != nil {
			return fmt.Errorf("failed to transfer leadership to %d: %w", freeSpareNode.ID, err)
		}
		// Recreate client to point to the new leader.
		client, err = c.clientGetter(ctx)
		if err != nil {
			return fmt.Errorf("failed to create dqlite client: %w", err)
		}
	}

	// Remove the node from the cluster. Retry as the leadership transfer might still be in progress.
	// For a large database this might take some time.
	return control.RetryFor(ctx, 10, func() error {
		if err := client.Remove(ctx, memberToRemove.ID); err != nil {
			time.Sleep(5 * time.Second)
			return fmt.Errorf("failed to remove node %v from dqlite cluster: %w", memberToRemove, err)
		}
		return nil
	})
}
