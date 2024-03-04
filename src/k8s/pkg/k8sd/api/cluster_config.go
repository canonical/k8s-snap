package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"

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

func validateConfig(oldConfig types.ClusterConfig, newConfig types.ClusterConfig) error {
	// If load-balancer, ingress or gateway gets enabled=true,
	// the request should fail if network.enabled is not true
	if !vals.OptionalBool(oldConfig.Ingress.Enabled, false) && vals.OptionalBool(newConfig.Ingress.Enabled, false) {
		if !vals.OptionalBool(newConfig.Network.Enabled, false) {
			return errors.New("network should be enabled before ingress can be enabled")
		}
	}

	if !vals.OptionalBool(oldConfig.Gateway.Enabled, false) && vals.OptionalBool(newConfig.Gateway.Enabled, false) {
		if !vals.OptionalBool(newConfig.Network.Enabled, false) {
			return errors.New("network should be enabled before gateway can be enabled")
		}
	}

	if !vals.OptionalBool(oldConfig.LoadBalancer.Enabled, false) && vals.OptionalBool(newConfig.LoadBalancer.Enabled, false) {
		if !vals.OptionalBool(newConfig.Network.Enabled, false) {
			return errors.New("network should be enabled before load-balancer can be enabled")
		}
	}

	// dns.service-ip should be in IP format and in service CIDR
	if newConfig.Kubelet.ClusterDNS != "" && net.ParseIP(newConfig.Kubelet.ClusterDNS) == nil {
		return errors.New("dns.service-ip needs to be in valid IP format")
	}

	// dns.service-ip is not changable if already dns.enabled=true.
	if vals.OptionalBool(newConfig.DNS.Enabled, false) && vals.OptionalBool(oldConfig.DNS.Enabled, false) {
		if newConfig.Kubelet.ClusterDNS != oldConfig.Kubelet.ClusterDNS {
			return errors.New("dns.service-ip can not be changed after dns is enabled")
		}
	}

	// load-balancer.bgp-enabled=true should fail if any of the bgp config is empty
	if vals.OptionalBool(newConfig.LoadBalancer.BGPEnabled, false) {
		if newConfig.LoadBalancer.BGPLocalASN == 0 {
			return errors.New("load-balancer.bgp-local-asn needs to be set before load-balancer.bgp-mode can be enabled")
		}
		if newConfig.LoadBalancer.BGPPeerAddress == "" {
			return errors.New("load-balancer.bgp-peer-address needs to be set before load-balancer.bgp-mode can be enabled")
		}
		if newConfig.LoadBalancer.BGPPeerPort == 0 {
			return errors.New("load-balancer.bgp-peer-port needs to be set before load-balancer.bgp-mode can be enabled")
		}
		if newConfig.LoadBalancer.BGPPeerASN == 0 {
			return errors.New("load-balancer.bgp-peer-asn needs to be set before load-balancer.bgp-mode can be enabled")
		}
	}

	// local-storage.local-path should not be changable if local-storage.enabled=true
	if vals.OptionalBool(newConfig.LocalStorage.Enabled, false) && vals.OptionalBool(oldConfig.LocalStorage.Enabled, false) {
		if newConfig.LocalStorage.LocalPath != oldConfig.LocalStorage.LocalPath {
			return errors.New("local-storage.local-path can not be changed after local-storage is enabled")
		}
	}

	// local-storage.reclaim-policy should be one of 3 values
	if !slices.Contains([]string{"Retain", "Recycle", "Delete"}, newConfig.LocalStorage.ReclaimPolicy) {
		return errors.New("local-storage.reclaim-policy should be one of: Retain, Recycle, Delete")
	}

	// local-storage.reclaim-policy should not be changable if local-storage.enabled=true
	if vals.OptionalBool(newConfig.LocalStorage.Enabled, false) && vals.OptionalBool(oldConfig.LocalStorage.Enabled, false) {
		if newConfig.LocalStorage.ReclaimPolicy != oldConfig.LocalStorage.ReclaimPolicy {
			return errors.New("local-storage.reclaim-policy can not be changed after local-storage is enabled")
		}
	}

	// network.enabled=false should not work before  load-balancer, ingress and gateway is disabled
	if vals.OptionalBool(oldConfig.Network.Enabled, false) && !vals.OptionalBool(newConfig.Network.Enabled, false) {
		if vals.OptionalBool(newConfig.Ingress.Enabled, false) {
			return errors.New("ingress should be disabled before network can be disabled")
		}
		if vals.OptionalBool(newConfig.Gateway.Enabled, false) {
			return errors.New("gateway should be disabled before network can be disabled")
		}
		if vals.OptionalBool(newConfig.LoadBalancer.Enabled, false) {
			return errors.New("load-balancer should be disabled before network can be disabled")
		}
	}

	return nil
}

func putClusterConfig(s *state.State, r *http.Request) response.Response {
	var req api.UpdateClusterConfigRequest
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	var oldConfig types.ClusterConfig

	if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		oldConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to read old cluster configuration: %w", err)
		}

		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to read cluster configuration failed: %w", err))
	}

	newConfig, err := types.MergeClusterConfig(oldConfig, types.ClusterConfigFromUserFacing(&req.Config))
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to merge new cluster config: %w", err))
	}

	if err := validateConfig(oldConfig, newConfig); err != nil {
		return response.InternalError(fmt.Errorf("config validation failed: %w", err))
	}

	if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		if err := database.SetClusterConfig(ctx, tx, newConfig); err != nil {
			return fmt.Errorf("failed to update cluster configuration: %w", err)
		}

		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
	}

	if req.Config.Network != nil {
		err := component.ReconcileNetworkComponent(r.Context(), snap, oldConfig.Network.Enabled, req.Config.Network.Enabled, newConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile network: %w", err))
		}
	}

	if req.Config.DNS != nil {
		dnsIP, _, err := component.ReconcileDNSComponent(r.Context(), snap, oldConfig.DNS.Enabled, req.Config.DNS.Enabled, newConfig)
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
		err := component.ReconcileLocalStorageComponent(r.Context(), snap, oldConfig.LocalStorage.Enabled, req.Config.LocalStorage.Enabled, newConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile local-storage: %w", err))
		}
	}

	if req.Config.Gateway != nil {
		err := component.ReconcileGatewayComponent(r.Context(), snap, oldConfig.Gateway.Enabled, req.Config.Gateway.Enabled, newConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile gateway: %w", err))
		}
	}

	if req.Config.Ingress != nil {
		err := component.ReconcileIngressComponent(r.Context(), snap, oldConfig.Ingress.Enabled, req.Config.Ingress.Enabled, newConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile ingress: %w", err))
		}
	}

	if req.Config.LoadBalancer != nil {
		err := component.ReconcileLoadBalancerComponent(r.Context(), snap, oldConfig.LoadBalancer.Enabled, req.Config.LoadBalancer.Enabled, newConfig)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to reconcile load-balancer: %w", err))
		}
	}

	if req.Config.MetricsServer != nil {
		err := component.ReconcileMetricsServerComponent(r.Context(), snap, oldConfig.MetricsServer.Enabled, req.Config.MetricsServer.Enabled, newConfig)
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
