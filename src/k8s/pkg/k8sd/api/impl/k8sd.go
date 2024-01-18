package impl

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/state"
)

// GetClusterStatus retrieves the status of the cluster, including information about its members.
func GetClusterStatus(ctx context.Context, s *state.State) (apiv1.ClusterStatus, error) {
	snap := snap.SnapFromContext(s.Context)

	members, err := GetClusterMembers(ctx, s)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster members: %w", err)
	}

	components, err := GetComponents(snap)
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to get cluster components: %w", err)
	}

	k8sClient, err := k8s.NewClient()
	if err != nil {
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to create k8s client: %w", err)
	}

	ready, err := k8s.ClusterReady(ctx, k8sClient)
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

// CreateK8sdToken creates a token entry in the k8sd db that can be used by a node to join.
func CreateK8sdToken(ctx context.Context, s *state.State, name string) (string, error) {
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

// GetClusterConfiguration queries the k8sd leader node for the k8s service configurations
// The k8sd token is used to authenticate this request.
func GetClusterConfiguration(ctx context.Context, s *state.State, k8sdToken string) (apiv1.JoinClusterResponse, error) {
	c, err := s.Leader()
	if err != nil {
		return apiv1.JoinClusterResponse{}, fmt.Errorf("failed to get leader client: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := apiv1.JoinClusterRequest{
		Token: k8sdToken,
	}

	var response apiv1.JoinClusterResponse
	err = c.Query(queryCtx, "POST", api.NewURL().Path("k8sd", "cluster", "join"), &request, &response)
	if err != nil {
		return apiv1.JoinClusterResponse{}, fmt.Errorf("failed to get cluster config: %w", err)
	}

	return response, nil
}
