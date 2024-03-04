package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/canonical/k8s/api/v1"

	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func putClusterConfig(s *state.State, r *http.Request) response.Response {
	var req api.UpdateClusterConfigRequest
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	var oldConfig types.ClusterConfig
	var clusterConfig types.ClusterConfig

	if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		oldConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to read old cluster configuration: %w", err)
		}

		if err := database.SetClusterConfig(ctx, tx, types.ClusterConfigFromUserFacing(&req.Config)); err != nil {
			return fmt.Errorf("failed to update cluster configuration: %w", err)
		}

		clusterConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to read new cluster configuration: %w", err)
		}

		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
	}

	if req.Config.Network != nil {
		err := component.ReconcileNetworkComponent(r.Context(), snap, oldConfig.Network.Enabled, req.Config.Network.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile network: %w", err))
		}
	}

	if req.Config.DNS != nil {
		dnsIP, _, err := component.ReconcileDNSComponent(r.Context(), snap, oldConfig.DNS.Enabled, req.Config.DNS.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile dns: %w", err))
		}
		if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			if err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: dnsIP,
				},
			}); err != nil {
				return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
			}
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
		}
	}

	if req.Config.LocalStorage != nil {
		err := component.ReconcileLocalStorageComponent(r.Context(), snap, oldConfig.LocalStorage.Enabled, req.Config.LocalStorage.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile local-storage: %w", err))
		}
	}

	if req.Config.Gateway != nil {
		err := component.ReconcileGatewayComponent(r.Context(), snap, oldConfig.Gateway.Enabled, req.Config.Gateway.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile gateway: %w", err))
		}
	}

	if req.Config.Ingress != nil {
		err := component.ReconcileIngressComponent(r.Context(), snap, oldConfig.Ingress.Enabled, req.Config.Ingress.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile ingress: %w", err))
		}
	}

	if req.Config.LoadBalancer != nil {
		err := component.ReconcileLoadBalancerComponent(r.Context(), snap, oldConfig.LoadBalancer.Enabled, req.Config.LoadBalancer.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	if req.Config.MetricsServer != nil {
		err := component.ReconcileMetricsServerComponent(r.Context(), snap, oldConfig.MetricsServer.Enabled, req.Config.MetricsServer.Enabled, clusterConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile metrics-server: %w", err))
		}
	}

	return response.SyncResponse(true, &api.UpdateClusterConfigResponse{})
}

func getClusterConfig(s *state.State, r *http.Request) response.Response {
	userFacing, err := utils.GetUserFacingClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get user-facing cluster config: %w", err))
	}

	result := api.GetClusterConfigResponse{
		Config: userFacing,
	}

	return response.SyncResponse(true, &result)
}
