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
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
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

	oldConfig, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
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
		if err := component.ReconcileNetworkComponent(r.Context(), snap, oldConfig.Network.Enabled, requestedConfig.Network.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile network: %w", err))
		}
	}

	if !requestedConfig.DNS.Empty() {
		dnsIP, _, err := component.ReconcileDNSComponent(r.Context(), snap, oldConfig.DNS.Enabled, requestedConfig.DNS.Enabled, mergedConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile dns: %w", err))
		}

		// If DNS IP is not empty, update cluster configuration
		if dnsIP != "" {
			if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
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

	cmData, err := mergedConfig.Kubelet.ToConfigMap()
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to format kubelet configmap data: %w", err))
	}

	client, err := k8s.NewClient(snap.KubernetesRESTClientGetter(""))
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create kubernetes client: %w", err))
	}

	if _, err := client.UpdateConfigMap(r.Context(), "kube-system", "k8sd-config", cmData); err != nil {
		return response.InternalError(fmt.Errorf("failed to update node config: %w", err))
	}

	if !requestedConfig.LocalStorage.Empty() {
		if err := component.ReconcileLocalStorageComponent(r.Context(), snap, oldConfig.LocalStorage.Enabled, requestedConfig.LocalStorage.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile local-storage: %w", err))
		}
	}

	if !requestedConfig.Gateway.Empty() {
		if err := component.ReconcileGatewayComponent(r.Context(), snap, oldConfig.Gateway.Enabled, requestedConfig.Gateway.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile gateway: %w", err))
		}
	}

	if !requestedConfig.Ingress.Empty() {
		if err := component.ReconcileIngressComponent(r.Context(), snap, oldConfig.Ingress.Enabled, requestedConfig.Ingress.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile ingress: %w", err))
		}
	}

	if !requestedConfig.LoadBalancer.Empty() {
		if err := component.ReconcileLoadBalancerComponent(r.Context(), snap, oldConfig.LoadBalancer.Enabled, requestedConfig.LoadBalancer.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	if !requestedConfig.MetricsServer.Empty() {
		if err := component.ReconcileMetricsServerComponent(r.Context(), snap, oldConfig.MetricsServer.Enabled, requestedConfig.MetricsServer.Enabled, mergedConfig); err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	return response.SyncResponse(true, &api.UpdateClusterConfigResponse{})
}

func (e *Endpoints) getClusterConfig(s *state.State, r *http.Request) response.Response {
	config, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}

	result := api.GetClusterConfigResponse{
		Config: config.ToUserFacing(),
	}
	return response.SyncResponse(true, &result)
}
