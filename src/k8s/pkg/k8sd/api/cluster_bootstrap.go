package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterBootstrap(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
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
	_, err = m.Status()
	if err == nil {
		return response.BadRequest(fmt.Errorf("cluster is already bootstrapped"))
	}

	// Bootstrap the cluster
	// Timeout 0 should leave the timeout to context via the m.ctx
	if err := m.NewCluster(hostname, req.Address, config, 0); err != nil {
		// TODO move node cleanup here
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	result := apiv1.NodeStatus{
		Name:    hostname,
		Address: req.Address,
	}

	return response.SyncResponse(true, &result)
}
