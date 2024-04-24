package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/utils"
	"net/http"

	api "github.com/canonical/k8s/api/v1"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) putClusterConfig(s *state.State, r *http.Request) response.Response {
	var req api.UpdateClusterConfigRequest
	snap := e.provider.Snap()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	requestedConfig, err := types.ClusterConfigFromUserFacing(req.Config)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid configuration: %w", err))
	}
	if requestedConfig.Datastore, err = types.DatastoreConfigFromUserFacing(req.Datastore); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse datastore config: %w", err))
	}

	var mergedConfig types.ClusterConfig
	if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		mergedConfig, err = database.SetClusterConfig(ctx, tx, requestedConfig)
		if err != nil {
			return fmt.Errorf("failed to update cluster configuration: %w", err)
		}

		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
	}

	if !requestedConfig.Network.Empty() {
		if err := features.ApplyNetwork(s.Context, snap, mergedConfig.Network); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply network: %w", err))
		}
	}

	if !requestedConfig.DNS.Empty() {
		dnsIP, err := features.ApplyDNS(s.Context, snap, mergedConfig.DNS, mergedConfig.Kubelet)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to apply DNS: %w", err))
		}

		if dnsIP != "" {
			if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
				if mergedConfig, err = database.SetClusterConfig(ctx, tx, types.ClusterConfig{
					Kubelet: types.Kubelet{
						ClusterDNS: utils.Pointer(dnsIP),
					},
				}); err != nil {
					return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
				}
				return nil
			}); err != nil {
				return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
			}
		}
	}

	if !requestedConfig.LocalStorage.Empty() {
		if err := features.ApplyLocalStorage(s.Context, snap, mergedConfig.LocalStorage); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply local-storage: %w", err))
		}
	}
	if !requestedConfig.Gateway.Empty() {
		if err := features.ApplyGateway(s.Context, snap, mergedConfig.Gateway); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply gateway: %w", err))
		}
	}
	if !requestedConfig.Ingress.Empty() {
		if err := features.ApplyIngress(s.Context, snap, mergedConfig.Ingress); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply ingress: %w", err))
		}
	}
	if !requestedConfig.LoadBalancer.Empty() {
		if err := features.ApplyLoadBalancer(s.Context, snap, mergedConfig.LoadBalancer); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply load-balancer: %w", err))
		}
	}
	if !requestedConfig.MetricsServer.Empty() {
		if err := features.ApplyMetricsServer(s.Context, snap, mergedConfig.MetricsServer); err != nil {
			return response.InternalError(fmt.Errorf("failed to apply metrics-server: %w", err))
		}
	}

	e.provider.NotifyUpdateConfigMap()

	return response.SyncResponse(true, &api.UpdateClusterConfigResponse{})
}

func (e *Endpoints) getClusterConfig(s *state.State, r *http.Request) response.Response {
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	result := api.GetClusterConfigResponse{
		Config: config.ToUserFacing(),
	}
	return response.SyncResponse(true, &result)
}
