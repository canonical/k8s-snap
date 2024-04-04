package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) getClusterStatus(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	members, err := impl.GetClusterMembers(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster members: %w", err))
	}

	config, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get user-facing cluster config: %w", err))
	}

	clusterConfig, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}
	datastoreConfig := apiv1.Datastore{
		Type:        *clusterConfig.Datastore.Type,
		ExternalURL: *clusterConfig.Datastore.ExternalURL,
	}

	client, err := k8s.NewClient(snap.KubernetesRESTClientGetter(""))
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	ready, err := client.HasReadyNodes(s.Context)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if cluster has ready nodes: %w", err))
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: apiv1.ClusterStatus{
			Ready:     ready,
			Members:   members,
			Config:    config.ToUserFacing(),
			Datastore: datastoreConfig,
		},
	}

	return response.SyncResponse(true, &result)
}
