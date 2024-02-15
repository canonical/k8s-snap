package utils

import (
	"context"
	"fmt"

	"github.com/canonical/microcluster/state"
)

// IsControlPlaneNode returns true if the given node name belongs to a control-plane in the cluster.
func IsControlPlaneNode(ctx context.Context, s *state.State, name string) (bool, error) {
	client, err := s.Leader()
	if err != nil {
		return false, fmt.Errorf("failed to get microcluster leader client: %w", err)
	}

	members, err := client.GetClusterMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get microcluster members: %w", err)
	}

	for _, member := range members {
		if member.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// IsWorkerNode returns true if the given node name belongs to a worker node in the cluster.
func IsWorkerNode(ctx context.Context, s *state.State, name string) (bool, error) {
	exists, err := CheckWorkerExists(ctx, s, name)
	if err != nil {
		return false, fmt.Errorf("failed to check if worker node %q exists: %w", name, err)
	}
	return exists, nil
}
