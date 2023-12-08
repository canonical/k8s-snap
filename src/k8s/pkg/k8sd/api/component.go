package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/k8sd/api/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"
)

var k8sdComponents = rest.Endpoint{
	Path: "k8sd/components",
	Get:  rest.EndpointAction{Handler: componentsGet, AllowUntrusted: false},
}

var k8sdComponentsName = rest.Endpoint{
	Path: "k8sd/components/{name}",
	Put:  rest.EndpointAction{Handler: componentsNamePut, AllowUntrusted: false},
}

func componentsGet(s *state.State, r *http.Request) response.Response {
	components, err := utils.GetComponents()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to get components: %w", err))
	}

	result := api.GetComponentsResponse{
		Components: components,
	}

	return response.SyncResponse(true, &result)
}

func componentsNamePut(s *state.State, r *http.Request) response.Response {
	componentName, err := url.PathUnescape(mux.Vars(r)["name"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse component name from URL '%s': %w", r.URL, err))
	}

	var req api.UpdateComponentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request: %w", err))
	}

	manager, err := component.NewManager()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to get component manager: %w", err))
	}

	if req.Status == api.ComponentEnable {
		err = manager.Enable(componentName)
	} else {
		err = manager.Disable(componentName)
	}

	if err != nil {
		return response.SmartError(fmt.Errorf("failed to %s %s: %w", req.Status, componentName, err))
	}

	return response.SyncResponse(true, &api.UpdateComponentResponse{})
}
