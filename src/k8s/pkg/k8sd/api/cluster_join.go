package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterJoin(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
	req := apiv1.JoinClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	// differentiate between control plane and worker node tokens
	info := &types.InternalWorkerNodeToken{}
	if info.Decode(req.Token) == nil {
		// valid worker node token
		if err := m.NewCluster(hostname, req.Address, map[string]string{"workerToken": req.Token}, time.Second*180); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as worker: %w", err))
		}
	} else {
		if err := m.JoinCluster(hostname, req.Address, req.Token, nil, time.Second*180); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as control plane: %w", err))
		}
	}

	return response.SyncResponse(true, nil)
}
