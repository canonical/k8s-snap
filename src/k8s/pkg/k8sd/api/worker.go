package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func postWorkerInfo(s *state.State, r *http.Request) response.Response {
	// TODO: move authentication through the HTTP token to an AccessHandler for the endpoint.
	token := r.Header.Get("k8sd-token")
	if token == "" {
		return response.Unauthorized(fmt.Errorf("invalid token"))
	}
	var tokenIsValid bool
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokenIsValid, err = database.CheckWorkerNodeToken(ctx, tx, token)
		if err != nil {
			return fmt.Errorf("failed to check worker node token: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("check token database transaction failed: %w", err))
	}
	if !tokenIsValid {
		return response.Unauthorized(fmt.Errorf("invalid token"))
	}

	req := apiv1.WorkerNodeInfoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}
	nodeName := req.Hostname
	if nodeName == "" {
		return response.BadRequest(fmt.Errorf("node name cannot be empty"))
	}
	nodeIP := net.ParseIP(req.Address)
	if nodeIP == nil {
		return response.BadRequest(fmt.Errorf("failed to parse node IP address %s", req.Address))
	}

	cfg, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	certificates := pki.NewControlPlanePKI("", nil, nil, 10, false)
	certificates.CACert = cfg.Certificates.CACert
	certificates.CAKey = cfg.Certificates.CAKey
	workerCertificates, err := certificates.CompleteWorkerNodePKI(nodeName, nodeIP, 2048)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to generate worker PKI: %w", err))
	}

	snap := snap.SnapFromContext(s.Context)
	client, err := k8s.NewClient(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create kubernetes client: %w", err))
	}
	servers, err := k8s.GetKubeAPIServerEndpoints(s.Context, client)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve list of known kube-apiserver endpoints: %w", err))
	}

	var (
		kubeletToken string
		proxyToken   string
	)
	for _, i := range []struct {
		token    *string
		name     string
		username string
		groups   []string
	}{
		{token: &kubeletToken, name: "kubelet", username: fmt.Sprintf("system:node:%s", nodeName), groups: []string{"system:nodes"}},
		{token: &proxyToken, name: "kube-proxy", username: "system:kube-proxy"},
	} {
		if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
			t, err := database.GetOrCreateToken(ctx, tx, i.username, i.groups)
			if err != nil {
				return fmt.Errorf("failed to generate %s token for node %q: %w", i.name, nodeName, err)
			}
			*i.token = t
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("create token transaction failed: %w", err))
		}
	}

	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.AddWorkerNode(ctx, tx, nodeName)
	}); err != nil {
		return response.InternalError(fmt.Errorf("add worker node transaction failed: %w", err))
	}

	return response.SyncResponse(true, &apiv1.WorkerNodeInfoResponse{
		CA:             cfg.Certificates.CACert,
		APIServers:     servers,
		PodCIDR:        cfg.Network.PodCIDR,
		KubeletToken:   kubeletToken,
		KubeProxyToken: proxyToken,
		ClusterDomain:  cfg.Kubelet.ClusterDomain,
		ClusterDNS:     cfg.Kubelet.ClusterDNS,
		CloudProvider:  cfg.Kubelet.CloudProvider,
		KubeletCert:    workerCertificates.KubeletCert,
		KubeletKey:     workerCertificates.KubeletKey,
	})
}
