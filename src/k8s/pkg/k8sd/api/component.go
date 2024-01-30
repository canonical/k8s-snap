package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/canonical/k8s/api/v1"

	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func getComponents(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

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
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableDNSComponent(snap, req.Config.ClusterDomain, req.Config.ServiceIP, req.Config.UpstreamNameservers); err != nil {
			return response.InternalError(fmt.Errorf("failed to enable dns: %w", err))
		}
	case api.ComponentDisable:
		if err := component.DisableDNSComponent(snap); err != nil {
			return response.InternalError(fmt.Errorf("failed to disable dns: %w", err))
		}
	default:
		return response.BadRequest(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func putNetworkComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateNetworkComponentRequest
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableNetworkComponent(snap); err != nil {
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
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableStorageComponent(snap); err != nil {
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
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableIngressComponent(snap, req.Config.DefaultTLSSecret, req.Config.EnableProxyProtocol); err != nil {
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
	snap := snap.SnapFromContext(s.Context)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		if err := component.EnableGatewayComponent(snap); err != nil {
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
