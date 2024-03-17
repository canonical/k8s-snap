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
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func putClusterConfig(s *state.State, r *http.Request) response.Response {
	var req api.UpdateClusterConfigRequest
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	oldConfig, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	requestedConfig := types.ClusterConfigFromUserFacing(&req.Config)
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

	if !requestedConfig.Features.Network.Empty() {
		if err := component.ReconcileNetworkComponent(r.Context(), snap, oldConfig.Features.Network.Enabled, requestedConfig.Features.Network.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile network: %w", err))
		}
	}

	if !requestedConfig.Features.DNS.Empty() {
		dnsIP, _, err := component.ReconcileDNSComponent(r.Context(), snap, oldConfig.Features.DNS.Enabled, requestedConfig.Features.DNS.Enabled, mergedConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile dns: %w", err))
		}

		// If DNS IP is not empty, update cluster configuration
		if dnsIP != "" {
			if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
				var err error
				mergedConfig, err = database.SetClusterConfig(ctx, tx, types.ClusterConfig{
					Kubelet: types.Kubelet{
						ClusterDNS: vals.Pointer(dnsIP),
					},
				})
				if err != nil {
					return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
				}
				return nil
			}); err != nil {
				return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
			}
		}
	}

	if !requestedConfig.Features.LocalStorage.Empty() {
		if err := component.ReconcileLocalStorageComponent(r.Context(), snap, oldConfig.Features.LocalStorage.Enabled, requestedConfig.Features.LocalStorage.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile local-storage: %w", err))
		}
	}

	if !requestedConfig.Features.Gateway.Empty() {
		if err := component.ReconcileGatewayComponent(r.Context(), snap, oldConfig.Features.Gateway.Enabled, requestedConfig.Features.Gateway.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile gateway: %w", err))
		}
	}

	if !requestedConfig.Features.Ingress.Empty() {
		if err := component.ReconcileIngressComponent(r.Context(), snap, oldConfig.Features.Ingress.Enabled, requestedConfig.Features.Ingress.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile ingress: %w", err))
		}
	}

	if !requestedConfig.Features.LoadBalancer.Empty() {
		if err := component.ReconcileLoadBalancerComponent(r.Context(), snap, oldConfig.Features.LoadBalancer.Enabled, requestedConfig.Features.LoadBalancer.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	if !requestedConfig.Features.MetricsServer.Empty() {
		if err := component.ReconcileMetricsServerComponent(r.Context(), snap, oldConfig.Features.MetricsServer.Enabled, requestedConfig.Features.MetricsServer.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	return response.SyncResponse(true, &api.UpdateClusterConfigResponse{})
}

func getClusterConfig(s *state.State, r *http.Request) response.Response {
	config, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	result := api.GetClusterConfigResponse{
		Config: types.ClusterConfigToUserFacing(config),
	}
	return response.SyncResponse(true, &result)
}
