package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterBootstrap(s *state.State, r *http.Request) response.Response {
	req := apiv1.PostClusterBootstrapRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	//Convert Bootstrap config to map
	config, err := req.Config.ToMicrocluster()
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to prepare bootstrap config: %w", err))
	}

	// Clean hostname
	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	// Check if the cluster is already bootstrapped
	_, err = e.provider.MicroCluster().Status(r.Context())
	if err == nil {
		return response.BadRequest(fmt.Errorf("cluster is already bootstrapped"))
	}

	// Bootstrap the cluster
	if err := e.provider.MicroCluster().NewCluster(r.Context(), hostname, req.Address, config); err != nil {
		// TODO move node cleanup here
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	result := apiv1.NodeStatus{
		Name:    hostname,
		Address: req.Address,
	}

	return response.SyncResponse(true, &result)
}
