package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeconfig(s *state.State, r *http.Request) response.Response {
	req := apiv1.GetKubeConfigRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}
	server := req.Server
	if req.Server == "" {
		server = fmt.Sprintf("https://%s:6443", s.Address().Hostname())
	}

	config := clientcmd.GetConfigFromFileOrDie("/etc/kubernetes/admin.conf")
	if _, ok := config.Clusters["k8s"]; !ok {
		return response.InternalError(fmt.Errorf("failed to read 'k8s' cluster data from kubeconfig"))
	}
	config.Clusters["k8s"].Server = server

	bConfig, err := clientcmd.Write(*config)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to serialize modified kubeconfig: %w", err))
	}

	result := apiv1.GetKubeConfigResponse{
		KubeConfig: string(bConfig),
	}
	return response.SyncResponse(true, &result)
}
