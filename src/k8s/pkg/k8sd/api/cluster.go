package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/api/utils"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdCluster = rest.Endpoint{
	Path: "k8sd/cluster",
	Get:  rest.EndpointAction{Handler: clusterGet, AllowUntrusted: false},
	Post: rest.EndpointAction{Handler: clusterPost, AllowUntrusted: false},
}

func clusterGet(s *state.State, r *http.Request) response.Response {
	status, err := utils.GetClusterStatus(r.Context(), s)
	if err != nil {
		response.InternalError(err)
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: status,
	}

	return response.SyncResponse(true, &result)
}

func clusterPost(s *state.State, r *http.Request) response.Response {
	err := setup.InitFolders()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup folders: %w", err))
	}

	err = setup.InitServiceArgs()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup service arguments: %w", err))
	}

	err = setup.InitContainerd()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize containerd: %w", err))
	}

	certMan, err := setup.InitCertificates()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup certificates: %w", err))
	}

	err = setup.InitKubeconfigs(r.Context(), s, certMan.CA)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to kubeconfig files: %w", err))
	}

	err = setup.InitKubeApiserver()
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize kube-apiserver: %w", err))
	}

	err = setup.InitPermissions(r.Context())
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup permissions: %w", err))
	}

	err = utils.WriteK8sDqliteCertInfoToK8sd(r.Context(), s)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to write k8s-dqlite cert to k8sd: %w", err))
	}

	err = snap.StartService(r.Context(), "k8s")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to start services: %w", err))
	}

	return response.SyncResponse(true, &apiv1.InitClusterResponse{})
}
