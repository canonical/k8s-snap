package dqlite

import (
	"context"
	"fmt"

	"github.com/canonical/go-dqlite/client"
	"github.com/canonical/k8s/pkg/snap"
)

// GetK8sDqliteClusterMembers queries the local k8s-dqlite datastore for its members.
//
// TODO:
// This should be done by using the go-dqlite client implementation.
// However, when I tried to use it the client connects, but returns an empty cluster member list.
func GetK8sDqliteClusterMembers(ctx context.Context, snap snap.Snap) ([]NodeInfo, error) {
	c, err := client.DefaultNodeStore(snap.CommonPath("/var/lib/k8s-dqlite/cluster.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s-dqlite datastore: %w", err)
	}
	internalMembers, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	clusterMembers := []NodeInfo{}
	for _, member := range internalMembers {
		clusterMembers = append(clusterMembers, NodeInfo{
			ID:      member.ID,
			Address: member.Address,
			Role:    member.Role.String(),
		})
	}
	return clusterMembers, nil
}
