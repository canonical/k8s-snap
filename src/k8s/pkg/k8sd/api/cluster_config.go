package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) putClusterConfig(s state.State, r *http.Request) response.Response {
	var req apiv1.SetClusterConfigRequest

	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	requestedConfig, err := types.ClusterConfigFromUserFacing(req.Config)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid configuration: %w", err))
	}
	if requestedConfig.Datastore, err = types.DatastoreConfigFromUserFacing(req.Datastore); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse datastore config: %w", err))
	}

	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, requestedConfig); err != nil {
			return fmt.Errorf("failed to update cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
	}

	e.provider.NotifyUpdateNodeConfigController()

	// Annotations can affect any feature (e.g. advertise-all-pools affects LoadBalancer).
	// Notify all feature controllers when annotations are present so they re-reconcile.
	hasAnnotations := len(requestedConfig.Annotations) > 0
	e.provider.NotifyFeatureController(
		!requestedConfig.Network.Empty() || hasAnnotations,
		!requestedConfig.Gateway.Empty() || hasAnnotations,
		!requestedConfig.Ingress.Empty() || hasAnnotations,
		!requestedConfig.LoadBalancer.Empty() || hasAnnotations,
		!requestedConfig.LocalStorage.Empty() || hasAnnotations,
		!requestedConfig.MetricsServer.Empty() || hasAnnotations,
		!requestedConfig.DNS.Empty() || !requestedConfig.Kubelet.Empty() || hasAnnotations,
	)

	return response.SyncResponse(true, &apiv1.SetClusterConfigResponse{})
}

func (e *Endpoints) getClusterConfig(s state.State, r *http.Request) response.Response {
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	return response.SyncResponse(true, &apiv1.GetClusterConfigResponse{
		Config:      config.ToUserFacing(),
		Datastore:   config.Datastore.ToUserFacing(),
		PodCIDR:     config.Network.PodCIDR,
		ServiceCIDR: config.Network.ServiceCIDR,
	})
}
