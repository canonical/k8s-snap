package dqlite

import (
	"github.com/canonical/go-dqlite/v2/client"
)

// NodeInfo is information about a node in the dqlite cluster.
type NodeInfo = client.NodeInfo

// Voter is the role for nodes that participate in the Raft quorum.
var Voter = client.Voter

// StandBy is the role for nodes that do not participate in quroum but replicate the database.
var StandBy = client.StandBy

// Spare is the role for nodes that do not participate in quroum and do not replicate the database.
var Spare = client.Spare
