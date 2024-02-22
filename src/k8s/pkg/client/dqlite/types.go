package dqlite

import (
	"github.com/canonical/go-dqlite/client"
)

// NodeInfo is information about a node in the dqlite cluster.
type NodeInfo = client.NodeInfo

// Voter is the role for nodes that participate in the Raft quorum.
var Voter = client.Voter
