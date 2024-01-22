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
			req.Config.ClusterDomain,
			req.Config.ServiceIP,
			req.Config.UpstreamNameservers,
			snap,
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	snap := snap.SnapFromContext(s.Context)

	err = component.EnableNetworkComponent(snap)

	return response.SyncResponse(true, &api.UpdateDNSComponentResponse{})
}
