package setup

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/dqlite"
	"github.com/canonical/microcluster/state"
)

// TODO(neoaggelos): this is not part of the cluster setup.
func LeaveK8sDqliteCluster(ctx context.Context, snap snap.Snap, state *state.State) error {
	clusterConfig, err := utils.GetClusterConfig(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	address := fmt.Sprintf("%s:%d", state.Address().Hostname(), clusterConfig.K8sDqlite.Port)

	members, err := dqlite.GetK8sDqliteClusterMembers(ctx, snap)
	if err != nil {
		return fmt.Errorf("failed to get cluster members: %w", err)
	}

	// TODO: handle case where node is leader but there are successors (e.g. use client.Transfer)
	if err := dqlite.IsLeaderWithoutSuccessor(ctx, members, address); err != nil {
		return fmt.Errorf("failed to leave cluster: %w", err)
	}

	// TODO: do not use the dqlite shell to remove the node.
	return utils.RunCommand(ctx, "/snap/k8s/current/k8s/wrappers/commands/dqlite", "k8s", fmt.Sprintf(".remove %s", address))
}
