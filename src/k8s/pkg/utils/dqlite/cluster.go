package dqlite

import (
	"context"
	"fmt"
)

// NodeInfo is a wrapper around the internal dqlite node struct.
type NodeInfo struct {
	ID      uint64
	Address string
	Role    string
}

// IsLeaderWithoutSuccessor returns an error if the node on the given address is the only voter in the cluster
func IsLeaderWithoutSuccessor(ctx context.Context, members []NodeInfo, address string) error {
	numVoters := 0
	isVoter := false
	for _, member := range members {
		if member.Role == "voter" {
			numVoters++
			if member.Address == address {
				isVoter = true
			}
		}
	}

	if numVoters == 1 && isVoter && len(members) > 1 {
		return fmt.Errorf("node is k8s-dqlite leader without successor")
	}
	return nil
}
