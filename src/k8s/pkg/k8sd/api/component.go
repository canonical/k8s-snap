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
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdComponents = rest.Endpoint{
	Path: "k8sd/components",
	Get:  rest.EndpointAction{Handler: componentsGet, AllowUntrusted: false},
}

var k8sdDNSComponent = rest.Endpoint{
	Path: "k8sd/components/dns",
	Put:  rest.EndpointAction{Handler: dnsComponentPut, AllowUntrusted: false},
}

var k8sdNetworkComponent = rest.Endpoint{
	Path: "k8sd/components/network",
	Put:  rest.EndpointAction{Handler: networkComponentPut, AllowUntrusted: false},
}

var k8sdStorageComponent = rest.Endpoint{
	Path: "k8sd/components/storage",
	Put:  rest.EndpointAction{Handler: storageComponentPut, AllowUntrusted: false},
}

var k8sdIngressComponent = rest.Endpoint{
	Path: "k8sd/components/ingress",
	Put:  rest.EndpointAction{Handler: ingressComponentPut, AllowUntrusted: false},
}

func componentsGet(s *state.State, r *http.Request) response.Response {
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

func dnsComponentPut(s *state.State, r *http.Request) response.Response {
	var req api.UpdateDNSComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	if req.Status == api.ComponentEnable {
		err = component.EnableDNSComponent(
			snap,
			req.Config.ClusterDomain,
			req.Config.ServiceIP,
			req.Config.UpstreamNameservers,
		)
	} else {
		err = component.DisableDNSComponent(snap)
	}
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to %s %s: %w", req.Status, "dns", err))
	}

	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func networkComponentPut(s *state.State, r *http.Request) response.Response {
	var req api.UpdateNetworkComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	if req.Status == api.ComponentEnable {
		err = component.EnableNetworkComponent(snap)
	} else {
		err = component.DisableNetworkComponent(snap)
	}
	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func storageComponentPut(s *state.State, r *http.Request) response.Response {
	var req api.UpdateStorageComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	if req.Status == api.ComponentEnable {
		err = component.EnableStorageComponent(snap)
	} else {
		err = component.DisableStorageComponent(snap)
	}
	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}

func ingressComponentPut(s *state.State, r *http.Request) response.Response {
	var req api.UpdateIngressComponentRequest
	snap := snap.SnapFromContext(s.Context)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	if req.Status == api.ComponentEnable {
		err = component.EnableIngressComponent(
			snap,
			req.Config.DefaultTLSSecret,
			req.Config.EnableProxyProtocol,
		)
	} else {
		err = component.DisableIngressComponent(snap)
	}
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to %s %s: %w", req.Status, "ingress", err))
	}

	return response.SyncResponse(true, &api.UpdateIngressComponentResponse{})
}
