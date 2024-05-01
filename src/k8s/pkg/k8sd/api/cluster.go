package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) getClusterStatus(s *state.State, r *http.Request) response.Response {
	// fail if node is not initialized yet
	if !s.Database.IsOpen() {
		return response.Unavailable(fmt.Errorf("daemon not yet initialized"))
	}

	members, err := impl.GetClusterMembers(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster members: %w", err))
	}
	config, err := databaseutil.GetClusterConfig(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	client, err := e.provider.Snap().KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	ready, err := client.HasReadyNodes(s.Context)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if cluster has ready nodes: %w", err))
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: apiv1.ClusterStatus{
			Ready:   ready,
			Members: members,
			Config:  config.ToUserFacing(),
			Datastore: apiv1.Datastore{
				Type:    config.Datastore.GetType(),
				Servers: config.Datastore.GetExternalServers(),
			},
		},
	}

	return response.SyncResponse(true, &result)
}
