package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
)

func (e *Endpoints) getKubeconfig(s state.State, r *http.Request) response.Response {
	req := apiv1.KubeConfigRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Fetch pieces needed to render an admin kubeconfig: ca, server, token
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster config: %w", err))
	}
	server := req.Server
	if req.Server == "" {
		server = fmt.Sprintf("%s:%d", s.Address().Hostname(), config.APIServer.GetSecurePort())
	}

	kubeconfig, err := setup.KubeconfigString(server, config.Certificates.GetCACert(), config.Certificates.GetAdminClientCert(), config.Certificates.GetAdminClientKey())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get kubeconfig: %w", err))
	}

	return response.SyncResponse(true, &apiv1.KubeConfigResponse{
		KubeConfig: kubeconfig,
	})
}
