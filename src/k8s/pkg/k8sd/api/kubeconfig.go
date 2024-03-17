package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func getKubeconfig(s *state.State, r *http.Request) response.Response {
	req := apiv1.GetKubeConfigRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Fetch pieces needed to render an admin kubeconfig: ca, server, token
	config, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster config: %w", err))
	}
	server := req.Server
	if req.Server == "" {
		server = fmt.Sprintf("%s:%d", s.Address().Hostname(), config.APIServer.GetSecurePort())
	}
	token, err := impl.GetOrCreateAuthToken(s.Context, s, "kubernetes-admin", []string{"system:masters"})
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get admin token: %w", err))
	}

	kubeconfig, err := setup.KubeconfigString(token, server, config.Certificates.GetCACert())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get kubeconfig: %w", err))
	}
	result := apiv1.GetKubeConfigResponse{
		KubeConfig: kubeconfig,
	}
	return response.SyncResponse(true, &result)
}
