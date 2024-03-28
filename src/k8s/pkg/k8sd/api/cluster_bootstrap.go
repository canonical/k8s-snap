package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterBootstrap(s *state.State, r *http.Request) response.Response {
	req := apiv1.PostClusterBootstrapRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	req.Config.SetDefaults()

	//Convert Bootstrap config to map
	config, err := req.Config.ToMap()
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to convert bootstrap config to map: %w", err))
	}

	// Clean hostname
	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	// Check if the cluster is already bootstrapped
	_, err = e.provider.MicroCluster().Status()
	if err == nil {
		return response.BadRequest(fmt.Errorf("cluster is already bootstrapped"))
	}

	// Set timeout
	timeout := utils.TimeoutFromCtx(s.Context, 30*time.Second)

	// Bootstrap the cluster
	if err := e.provider.MicroCluster().NewCluster(hostname, req.Address, config, timeout); err != nil {
		// TODO move node cleanup here
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	result := apiv1.NodeStatus{
		Name:    hostname,
		Address: req.Address,
	}

	return response.SyncResponse(true, &result)
}
