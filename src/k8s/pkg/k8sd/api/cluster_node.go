package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"
)

func postClusterNode(s *state.State, r *http.Request) response.Response {
	nodeName, err := url.PathUnescape(mux.Vars(r)["node"])
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse node name from URL '%s': %w", r.URL, err))
	}
	req := apiv1.JoinNodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// differentiate between control plane and worker node tokens
	info := &types.InternalWorkerNodeToken{}
	if info.Decode(req.Token) == nil {
		// valid worker node token
		if err := joinWorkerNode(s, r, nodeName, req.Address, req.Token); err != nil {
			return response.SmartError(fmt.Errorf("failed to join k8sd cluster as worker: %w", err))
		}
	} else {
		if err := joinControlPlaneNode(s, r, nodeName, req.Address, req.Token); err != nil {
			return response.SmartError(fmt.Errorf("failed to join k8sd cluster as control plane: %w", err))
		}
	}

	return response.SyncResponse(true, &apiv1.JoinNodeResponse{})
}

func joinWorkerNode(s *state.State, r *http.Request, name, address, token string) error {
	m, err := microcluster.App(r.Context(), microcluster.Args{
		StateDir: s.OS.StateDir,
	})
	if err != nil {
		return fmt.Errorf("failed to get microcluster app: %w", err)
	}
	return m.NewCluster(name, address, map[string]string{"workerToken": token}, time.Second*180)
}

func joinControlPlaneNode(s *state.State, r *http.Request, name, address, token string) error {
	m, err := microcluster.App(r.Context(), microcluster.Args{
		StateDir: s.OS.StateDir,
	})
	if err != nil {
		return fmt.Errorf("failed to get microcluster app: %w", err)
	}
	return m.JoinCluster(name, address, token, nil, time.Second*180)
}
