package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/canonical/k8s/api/v1"

	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func getComponents(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(r.Context())

	components, err := impl.GetComponents(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get components: %w", err))
	}

	result := api.GetComponentsResponse{
		Components: components,
	}
	return response.SyncResponse(true, &result)
}

func putDNSComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateDNSComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		dnsIP, clusterDomain, err := component.EnableDNSComponent(r.Context(), snap, req.Config.ClusterDomain, req.Config.ServiceIP, req.Config.UpstreamNameservers)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to enable dns: %w", err))
		}

		if err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			if err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS:    dnsIP,
					ClusterDomain: clusterDomain,
				},
			}); err != nil {
				return fmt.Errorf("failed to update cluster configuration for dns=%s domain=%s: %w", dnsIP, clusterDomain, err)
			}
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("database transaction to update cluster configuration failed: %w", err))
		}

	case api.ComponentDisable:
		if err := component.DisableDNSComponent(r.Context(), snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable dns: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func putNetworkComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateNetworkComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		cfg, err := utils.GetClusterConfig(r.Context(), s)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to retrieve pod cidr: %w", err))
		}
		if err := component.EnableNetworkComponent(r.Context(), snap, cfg.Network.PodCIDR); err != nil {
			return response.InternalError(fmt.Errorf("failed to enable network: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableNetworkComponent(snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable network: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateNetworkComponentResponse{})
}

func putStorageComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateStorageComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableStorageComponent(r.Context(), snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to enable storage: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableStorageComponent(snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable storage: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateStorageComponentResponse{})
}

func putIngressComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateIngressComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableIngressComponent(r.Context(), snap, req.Config.DefaultTLSSecret, req.Config.EnableProxyProtocol); err != nil {
			return response.InternalError(fmt.Errorf("failed to enable ingress: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableIngressComponent(snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable ingress: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateIngressComponentResponse{})
}

func putGatewayComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateGatewayComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableGatewayComponent(r.Context(), snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to enable gateway API: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableGatewayComponent(snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable gateway API: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateGatewayComponentResponse{})
}

func putLoadBalancerComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateLoadBalancerComponentRequest
	snap := snap.SnapFromContext(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableLoadBalancerComponent(
			r.Context(),
			snap,
			req.Config.CIDRs,
			req.Config.L2Enabled,
			req.Config.L2Interfaces,
			req.Config.BGPEnabled,
			req.Config.BGPLocalASN,
			req.Config.BGPPeerAddress,
			req.Config.BGPPeerASN,
			req.Config.BGPPeerPort,
		); err != nil {
			return response.SmartError(fmt.Errorf("failed to enable loadbalancer: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableLoadBalancerComponent(snap); err != nil {
			return response.SmartError(fmt.Errorf("failed to disable loadbalancer: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateLoadBalancerComponentResponse{})
}
