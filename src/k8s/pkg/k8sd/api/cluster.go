package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
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
	var req apiv1.GetClusterStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to decode request data: %w", err))
	}

	status, err := impl.GetClusterStatus(r.Context(), s)
	if err != nil {
		response.InternalError(err)
	}

	result := apiv1.GetClusterStatusResponse{
		ClusterStatus: status,
	}

	return response.SyncResponse(true, &result)
}

func clusterPost(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

	err := setup.InitFolders(snap.DataPath("args"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup folders: %w", err))
	}

	err = setup.InitServiceArgs(snap, apiv1.ExtraServiceArgs{})
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup service arguments: %w", err))
	}

	err = setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize containerd: %w", err))
	}

	certMan, err := setup.InitCertificates(nil)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup certificates: %w", err))
	}

	err = setup.InitKubeconfigs(r.Context(), s, certMan.CA, nil, nil)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to kubeconfig files: %w", err))
	}

	err = setup.InitKubeApiserver(snap.Path("k8s/config/apiserver-token-hook.tmpl"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to initialize kube-apiserver: %w", err))
	}

	err = setup.InitPermissions(r.Context(), snap)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to setup permissions: %w", err))
	}

	err = cert.WriteCertKeyPairToK8sd(r.Context(), s, "k8s-dqlite",
		path.Join(cert.K8sDqlitePkiPath, "cluster.crt"), path.Join(cert.K8sDqlitePkiPath, "cluster.key"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to write k8s-dqlite cert to k8sd: %w", err))
	}

	err = cert.WriteCertKeyPairToK8sd(r.Context(), s, "ca",
		path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key"))
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to write CA to k8sd: %w", err))
	}

	err = snap.StartService(r.Context(), "k8s")
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to start services: %w", err))
	}

	return response.SyncResponse(true, &apiv1.InitClusterResponse{})
}
