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
		return response.SmartError(fmt.Errorf("failed to get components: %w", err))
	}

	result := api.GetComponentsResponse{
		Components: components,
	}

	return response.SyncResponse(true, &result)
}

func putDNSComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateDNSComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		err = component.EnableDNSComponent(
			snap,
			req.Config.ClusterDomain,
			req.Config.ServiceIP,
			req.Config.UpstreamNameservers,
		)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to enable dns: %w", err))
		}
	case api.ComponentDisable:
		err = component.DisableDNSComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to disable dns: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func putNetworkComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateNetworkComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		err = component.EnableNetworkComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to enable network: %w", err))
		}
	case api.ComponentDisable:
		err = component.DisableNetworkComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to disable network: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateNetworkComponentResponse{})
}

func putStorageComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateStorageComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		err = component.EnableStorageComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to enable storage: %w", err))
		}
	case api.ComponentDisable:
		err = component.DisableStorageComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to disable storage: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateStorageComponentResponse{})
}

func putIngressComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateIngressComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		err = component.EnableIngressComponent(
			snap,
			req.Config.DefaultTLSSecret,
			req.Config.EnableProxyProtocol,
		)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to enable ingress: %w", err))
		}
	case api.ComponentDisable:
		err = component.DisableIngressComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to disable ingress: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateIngressComponentResponse{})
}

func putGatewayComponent(s *state.State, r *http.Request) response.Response {
	var req api.UpdateGatewayComponentRequest

	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	switch req.Status {
	case api.ComponentEnable:
		err = component.EnableGatewayComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to enable gateway: %w", err))
		}
	case api.ComponentDisable:
		err = component.DisableGatewayComponent(snap)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to disable gateway: %w", err))
		}
	default:
		return response.SmartError(fmt.Errorf("invalid component status %s", req.Status))
	}

	return response.SyncResponse(true, &api.UpdateGatewayComponentResponse{})
}
