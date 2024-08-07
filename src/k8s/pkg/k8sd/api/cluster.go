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
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) getClusterStatus(s state.State, r *http.Request) response.Response {
	// fail if node is not initialized yet
	if err := s.Database().IsOpen(r.Context()); err != nil {
		return response.Unavailable(fmt.Errorf("daemon not yet initialized"))
	}

	members, err := impl.GetClusterMembers(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster members: %w", err))
	}
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	client, err := e.provider.Snap().KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	ready, err := client.HasReadyNodes(r.Context())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if cluster has ready nodes: %w", err))
	}

	var statuses map[string]types.FeatureStatus
	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		statuses, err = database.GetFeatureStatuses(r.Context(), tx)
		if err != nil {
			return fmt.Errorf("failed to get feature statuses: %w", err)
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
			DNS:           statuses["dns"].ToAPI(),
			Network:       statuses["network"].ToAPI(),
			LoadBalancer:  statuses["load-balancer"].ToAPI(),
			Ingress:       statuses["ingress"].ToAPI(),
			Gateway:       statuses["gateway"].ToAPI(),
			MetricsServer: statuses["metrics-server"].ToAPI(),
			LocalStorage:  statuses["local-storage"].ToAPI(),
		},
	}

	return response.SyncResponse(true, &result)
}
