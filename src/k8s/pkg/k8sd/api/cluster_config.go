package api

import (
	"fmt"
	"net/http"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func getKubeconfig(s *state.State, r *http.Request) response.Response {
	// TODO: Render a new kubeconfig instead of reading the existing one
	//       when the config can be altered via request parameters.
	config, err := os.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read admin kubeconfig: %w", err))
	}

	result := apiv1.GetKubeConfigResponse{
		KubeConfig: string(config),
	}
	return response.SyncResponse(true, &result)
}
