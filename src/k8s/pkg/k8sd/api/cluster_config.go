package api

import (
	"fmt"
	"net/http"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdClusterConfig = rest.Endpoint{
	Path: "k8sd/config",
	Get:  rest.EndpointAction{Handler: clusterConfigGet, AllowUntrusted: false},
}

func clusterConfigGet(s *state.State, r *http.Request) response.Response {
	config, err := os.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read admin kubeconfig: %w", err))
	}

	result := apiv1.GetKubeConfigResponse{
		KubeConfig: string(config),
	}
	return response.SyncResponse(true, &result)
}
