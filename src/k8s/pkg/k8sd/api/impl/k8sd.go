package impl

import (
	"context"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcluster/state"
)

// GetClusterStatus retrieves the status of the cluster, including information about its members.
func GetClusterStatus(ctx context.Context, s *state.State) (apiv1.ClusterStatus, error) {
	members, err := GetClusterMembers(ctx, s)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster members: %w", err)
	}

	components, err := GetComponents()
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster components: %w", err)
	}

	return apiv1.ClusterStatus{
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

// DeleteClusterMember deletes a member from the cluster.
func DeleteClusterMember(ctx context.Context, s *state.State, name string, force bool) error {
	c, err := s.Leader()
	if err != nil {
		return fmt.Errorf("failed to get leader client: %w", err)
	}

	err = c.DeleteClusterMember(ctx, name, force)
	if err != nil {
		return fmt.Errorf("failed to delete cluster member %q (forced=%t): %w", name, force, err)
	}

	return nil
}

// CreateJoinToken creates a token entry in the k8sd db that can be used by a node to join.
func CreateJoinToken(ctx context.Context, s *state.State, name string) (string, error) {
	c, err := s.Leader()
	if err != nil {
		return "", fmt.Errorf("failed to get leader client: %w", err)
	}

	token, err := c.RequestToken(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	return token, nil
}
