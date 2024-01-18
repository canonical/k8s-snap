package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	snapPkg "github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdClusterJoin = rest.Endpoint{
	Path: "k8sd/cluster/join",
	Post: rest.EndpointAction{Handler: clusterJoinPost, AllowUntrusted: false},
}

func clusterJoinPost(s *state.State, r *http.Request) response.Response {
	snap := snapPkg.SnapFromContext(s.Context)

	var req apiv1.AddNodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request data: %w", err))
	}

	k8sdToken, err := impl.K8sdTokenFromBase64Token(req.Token)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to parse token information: %w", err))
	}

	isValid, err := impl.CheckK8sdToken(r.Context(), s, k8sdToken)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to verify token: %w", err))
	}

	if !isValid {
		return response.SmartError(fmt.Errorf("token is not valid"))
	}

	apiServerPort := snapPkg.GetServiceArgument(snap, "kube-apiserver", "--secure-port")
	clusterCIDR := snapPkg.GetServiceArgument(snap, "kube-proxy", "--cluster-cidr")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read arguments of kubelet service: %w", err))
	}

	overwrites := map[string]map[string]string{
		"kube-apiserver": {
			"--secure-port": apiServerPort,
		},
		"kube-proxy": {
			"--cluster-cidr": clusterCIDR,
		},
	}

	return response.SyncResponse(true, &apiv1.JoinClusterResponse{
		ExtraServiceArgs: overwrites,
	})
}
