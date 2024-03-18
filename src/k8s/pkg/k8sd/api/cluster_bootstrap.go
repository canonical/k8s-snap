package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterBootstrap(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
	// TODO: set bootstrap config defaults?
	// bootstrapConfig := apiv1.BootstrapConfig{}
	// bootstrapConfig.SetDefaults()

	req := apiv1.ClusterBootstrapRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	//Convert Bootstrap config to map
	config, err := req.BootstrapConfig.ToMap()
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to convert bootstrap config to map: %w", err))
	}

	// Clean hostname
	// TODO: Move this to the server side
	hostname, err := utils.CleanHostname(s.Name())
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", s.Name(), err))
	}

	// Set timeout
	timeout := 30 * time.Second
	if deadline, set := s.Context.Deadline(); set {
		timeout = time.Until(deadline)
	}

	if err := m.Ready(int(timeout / time.Second)); err != nil {
		return response.BadRequest(fmt.Errorf("cluster did not come up in time: %w", err))
	}

	// TODO Check if already Bootstrapped

	// Bootstrap the cluster
	address := util.CanonicalNetworkAddress(util.NetworkInterfaceAddress(), config.DefaultPort)
	if err := m.NewCluster(hostname, address, config, timeout); err != nil {
		// TODO(neoaggelos): only return error that bootstrap failed
		fmt.Fprintln(os.Stderr, "Failed with error:", err)
		// c.CleanupNode(ctx, hostname)
		return response.BadRequest(fmt.Errorf("failed to bootstrap new cluster: %w", err))
	}

	return response.SyncResponse(true, nil)
}
