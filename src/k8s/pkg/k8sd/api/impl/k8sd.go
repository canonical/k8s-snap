package impl

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcluster/state"
)

// GetClusterStatus retrieves the status of the cluster, including information about its members.
func GetClusterStatus(ctx context.Context, s *state.State) (apiv1.ClusterStatus, error) {
	snap := snap.SnapFromContext(s.Context)

	k8sClient, err := k8s.NewClient(snap)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to create k8s client: %w", err)
	}

	err = k8s.WaitApiServerReady(ctx, k8sClient)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("k8s api server did not become ready in time: %w", err)
	}

	ready, err := k8s.ClusterReady(ctx, k8sClient)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster components: %w", err)
	}

	members, err := GetClusterMembers(ctx, s)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster members: %w", err)
	}

	components, err := GetComponents(snap)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster components: %w", err)
	}

	return apiv1.ClusterStatus{
		Ready:      ready,
		Members:    members,
		Components: components,
	}, nil
}

// GetClusterMembers retrieves information about the members of the cluster.
func GetClusterMembers(ctx context.Context, s *state.State) ([]apiv1.ClusterMember, error) {
	c, err := s.Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := c.GetClusterMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	members := make([]apiv1.ClusterMember, len(clusterMembers))
	for i, clusterMember := range clusterMembers {
		fingerprint, err := shared.CertFingerprintStr(clusterMember.Certificate.String())
		if err != nil {
			continue
		}

		members[i] = apiv1.ClusterMember{
			Name:        clusterMember.Name,
			Address:     clusterMember.Address.String(),
			Role:        clusterMember.Role,
			Fingerprint: fingerprint,
			Status:      string(clusterMember.Status),
		}
	}

	return members, nil
}
