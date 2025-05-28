package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) getClusterStatus(s state.State, r *http.Request) response.Response {
	// fail if node is not initialized yet

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - checking if database is open")

	if err := s.Database().IsOpen(r.Context()); err != nil {
		return response.Unavailable(fmt.Errorf("daemon not yet initialized"))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - database is open, proceeding to get cluster status")

	members, err := impl.GetClusterMembers(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster members: %w", err))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - got cluster members:", members)
	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - getting cluster config")

	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - got cluster config")

	client, err := e.provider.Snap().KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - created k8s client")

	ready, err := client.HasReadyNodes(r.Context())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if cluster has ready nodes: %w", err))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - checked if cluster has ready nodes:", ready)

	var statuses map[types.FeatureName]types.FeatureStatus
	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - getting feature statuses from database")
		statuses, err = database.GetFeatureStatuses(r.Context(), tx)
		if err != nil {
			return fmt.Errorf("failed to get feature statuses: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
	}

	log.Println("HUE - k8s-snap - cluster.go/getClusterStatus - got feature statuses:", statuses)

	return response.SyncResponse(true, &apiv1.ClusterStatusResponse{
		ClusterStatus: apiv1.ClusterStatus{
			Ready:   ready,
			Members: members,
			Config:  config.ToUserFacing(),
			Datastore: apiv1.Datastore{
				Type:    config.Datastore.GetType(),
				Servers: config.Datastore.GetExternalServers(),
			},
			DNS:           statuses[features.DNS].ToAPI(),
			Network:       statuses[features.Network].ToAPI(),
			LoadBalancer:  statuses[features.LoadBalancer].ToAPI(),
			Ingress:       statuses[features.Ingress].ToAPI(),
			Gateway:       statuses[features.Gateway].ToAPI(),
			MetricsServer: statuses[features.MetricsServer].ToAPI(),
			LocalStorage:  statuses[features.LocalStorage].ToAPI(),
		},
	})
}
