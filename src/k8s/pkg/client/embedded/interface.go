package embedded

import "context"

// Client handles the interaction with an embedded cluster database.
type Client interface {
	// RemoveNodeByAddress removes the member with the specified name from the cluster.
	RemoveNodeByAddress(ctx context.Context, peerURL string) error
}
