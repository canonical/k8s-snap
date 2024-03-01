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
	cfg, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	userFacing := api.UserFacingClusterConfig{
		Network: &api.NetworkConfig{
			Enabled: vals.Pointer(true),
		},
		DNS: &api.DNSConfig{
			Enabled:             vals.Pointer(true),
			UpstreamNameservers: cfg.DNS.UpstreamNameservers,
			ServiceIP:           cfg.Kubelet.ClusterDNS,
			ClusterDomain:       cfg.Kubelet.ClusterDomain,
		},
		Ingress: &api.IngressConfig{
			Enabled:             vals.Pointer(false),
			DefaultTLSSecret:    cfg.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: vals.Pointer(false),
		},
		LoadBalancer: &api.LoadBalancerConfig{
			Enabled:        vals.Pointer(false),
			L2Enabled:      vals.Pointer(false),
			L2Interfaces:   cfg.LoadBalancer.L2Interfaces,
			BGPEnabled:     vals.Pointer(false),
			BGPLocalASN:    cfg.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: cfg.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     cfg.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    cfg.LoadBalancer.BGPPeerPort,
		},
		LocalStorage: &api.LocalStorageConfig{
			Enabled:       vals.Pointer(false),
			LocalPath:     cfg.LocalStorage.LocalPath,
			ReclaimPolicy: cfg.LocalStorage.ReclaimPolicy,
			SetDefault:    vals.Pointer(true),
		},
		Gateway: &api.GatewayConfig{
			Enabled: vals.Pointer(false),
		},
		MetricsServer: &api.MetricsServerConfig{
			Enabled: vals.Pointer(false),
		},
	}

	if cfg.Network.Enabled != nil {
		userFacing.Network.Enabled = cfg.Network.Enabled
	}

	if cfg.DNS.Enabled != nil {
		userFacing.DNS.Enabled = cfg.DNS.Enabled
	}

	if cfg.Ingress.Enabled != nil {
		userFacing.Ingress.Enabled = cfg.Ingress.Enabled
	}

	if cfg.LoadBalancer.Enabled != nil {
		userFacing.LoadBalancer.Enabled = cfg.LoadBalancer.Enabled
	}

	if cfg.LocalStorage.Enabled != nil {
		userFacing.LocalStorage.Enabled = cfg.LocalStorage.Enabled
	}

	if cfg.Gateway.Enabled != nil {
		userFacing.Gateway.Enabled = cfg.Gateway.Enabled
	}

	if cfg.MetricsServer.Enabled != nil {
		userFacing.MetricsServer.Enabled = cfg.MetricsServer.Enabled
	}

	if cfg.Ingress.EnableProxyProtocol != nil {
		userFacing.Ingress.EnableProxyProtocol = cfg.Ingress.EnableProxyProtocol
	}

	if cfg.LoadBalancer.L2Enabled != nil {
		userFacing.LoadBalancer.L2Enabled = cfg.LoadBalancer.L2Enabled
	}

	if cfg.LoadBalancer.BGPEnabled != nil {
		userFacing.LoadBalancer.BGPEnabled = cfg.LoadBalancer.BGPEnabled
	}

	if cfg.LocalStorage.SetDefault != nil {
		userFacing.LocalStorage.SetDefault = cfg.LocalStorage.SetDefault
	}

	result := api.GetClusterConfigResponse{
		Config: userFacing,
	}

	return response.SyncResponse(true, &result)
}
