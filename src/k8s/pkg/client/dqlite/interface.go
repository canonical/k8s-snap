package dqlite

import (
	"context"

	"github.com/canonical/go-dqlite/client"
)

// NodeInfo is information about a node in the dqlite cluster.
type NodeInfo = client.NodeInfo

// Voter is the role for nodes that participate in the Raft quorum.
var Voter = client.Voter

// Interface wraps go-dqlite client methods.
type Interface interface {
	// Cluster returns a list of the members of the dqlite cluster.
	Cluster(ctx context.Context) ([]NodeInfo, error)
	// Remove removes a node from the dqlite cluster given its ID.
	Remove(ctx context.Context, id uint64) error
}

var _ Interface = &client.Client{}
