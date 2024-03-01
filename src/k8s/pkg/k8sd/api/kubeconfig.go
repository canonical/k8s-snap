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
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeconfig(s *state.State, r *http.Request) response.Response {
	req := apiv1.GetKubeConfigRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Fetch pieces needed to render an admin kubeconfig: ca, server, token
	clusterCfg, err := utils.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster config: %w", err))
	}
	ca := clusterCfg.Certificates.CACert
	server := req.Server
	if req.Server == "" {
		server = fmt.Sprintf("%s:%d", s.Address().Hostname(), clusterCfg.APIServer.SecurePort)
	}
	token, err := impl.GetOrCreateAuthToken(s.Context, s, "kubernetes-admin", []string{"system:masters"})
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get admin token: %w", err))
	}

	// Convert rendered kubeconfig bytes into yaml with help from clientcmd
	kubeconfig, err := setup.ClientKubeconfig(token, server, ca)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get kubeconfig: %w", err))
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create ClientConfig: %w", err))
	}
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create RawConfig: %w", err))
	}
	yamlConfig, err := clientcmd.Write(rawConfig)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to serialize modified kubeconfig: %w", err))
	}

	result := apiv1.GetKubeConfigResponse{
		KubeConfig: string(yamlConfig),
	}
	return response.SyncResponse(true, &result)
}
