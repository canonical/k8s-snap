package impl

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	nodeutil "github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/microcluster/state"
)

// GetClusterMembers retrieves information about the members of the cluster.
func GetClusterMembers(ctx context.Context, s *state.State) ([]apiv1.NodeStatus, error) {
	c, err := s.Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := c.GetClusterMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	members := make([]apiv1.NodeStatus, len(clusterMembers))
	for i, clusterMember := range clusterMembers {
		members[i] = apiv1.NodeStatus{
			Name:          clusterMember.Name,
			Address:       clusterMember.Address.String(),
			ClusterRole:   apiv1.ClusterRoleControlPlane,
			DatastoreRole: nodeutil.DatastoreRoleFromString(clusterMember.Role),
		}
	}

	return members, nil
}

// GetLocalNodeStatus retrieves the status of the local node, including its roles within the cluster.
// Unlike "GetClusterMembers" this also works on a worker node.
func GetLocalNodeStatus(ctx context.Context, s *state.State, snap snap.Snap) (apiv1.NodeStatus, error) {
	// Determine cluster role.
	clusterRole := apiv1.ClusterRoleUnknown
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to check if node is a worker: %w", err)
	}

	if isWorker {
		clusterRole = apiv1.ClusterRoleWorker
	} else if node, err := nodeutil.GetControlPlaneNode(ctx, s, s.Name()); err != nil {
		clusterRole = apiv1.ClusterRoleUnknown
	} else if node != nil {
		return *node, nil
	}

	return apiv1.NodeStatus{
		Name:        s.Name(),
		Address:     s.Address().Hostname(),
		ClusterRole: clusterRole,
	}, nil

}
