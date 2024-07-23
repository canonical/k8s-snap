package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/database"
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

	featureStatuses := make(map[string]apiv1.FeatureStatus)
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		statuses, err := database.GetFeatureStatuses(s.Context, tx)
		if err != nil {
			return fmt.Errorf("failed to get feature statuses: %w", err)
		}

		for name, st := range statuses {
			apiSt, err := st.ToAPI()
			if err != nil {
				return fmt.Errorf("failed to convert k8sd feature status to api feature status: %w", err)
			}

			featureStatuses[name] = apiSt
		}

		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
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
			DNS:           featureStatuses["dns"],
			Network:       featureStatuses["network"],
			LoadBalancer:  featureStatuses["load-balancer"],
			Ingress:       featureStatuses["ingress"],
			Gateway:       featureStatuses["gateway"],
			MetricsServer: featureStatuses["metrics-server"],
			LocalStorage:  featureStatuses["local-storage"],
		},
	}

	return response.SyncResponse(true, &result)
}
