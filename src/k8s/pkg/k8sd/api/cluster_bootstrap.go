package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterBootstrap(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
	req := apiv1.ClusterBootstrapRequest{}
	req.BootstrapConfig.SetDefaults()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	//Convert Bootstrap config to map
	config, err := req.BootstrapConfig.ToMap()
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to convert bootstrap config to map: %w", err))
	}

	// Clean hostname
	hostname, err := utils.CleanHostname(s.Name())
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", s.Name(), err))
	}

	// Check if the cluster is already bootstrapped
	_, err = m.Status()
	if err == nil {
		return response.BadRequest(fmt.Errorf("cluster is already bootstrapped"))
	}

	// Bootstrap the cluster
	address := util.CanonicalNetworkAddress(util.NetworkInterfaceAddress(), req.BootstrapConfig.K8sDqlitePort)
	if err := m.NewCluster(hostname, address, config, timeout); err != nil {
		// TODO move node cleanup here
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	return response.SyncResponse(true, nil)
}
