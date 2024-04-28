package api

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postWorkerInfo(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.WorkerNodeInfoRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Existence of this header is already checked in the access handler.
	workerName := r.Header.Get("worker-name")
	nodeIP := net.ParseIP(req.Address)
	if nodeIP == nil {
		return response.BadRequest(fmt.Errorf("failed to parse node IP address %s", req.Address))
	}

	cfg, err := databaseutil.GetClusterConfig(s.Context, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{Years: 10})
	certificates.CACert = cfg.Certificates.GetCACert()
	certificates.CAKey = cfg.Certificates.GetCAKey()
	workerCertificates, err := certificates.CompleteWorkerNodePKI(workerName, nodeIP, 2048)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to generate worker PKI: %w", err))
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create kubernetes client: %w", err))
	}
	if err := client.WaitApiServerReady(s.Context); err != nil {
		return response.InternalError(fmt.Errorf("kube-apiserver did not become ready in time: %w", err))
	}
	servers, err := client.GetKubeAPIServerEndpoints(s.Context)
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
		{token: &kubeletToken, name: "kubelet", username: fmt.Sprintf("system:node:%s", workerName), groups: []string{"system:nodes"}},
		{token: &proxyToken, name: "kube-proxy", username: "system:kube-proxy"},
	} {
		if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
			t, err := database.GetOrCreateToken(ctx, tx, i.username, i.groups)
			if err != nil {
				return fmt.Errorf("failed to generate %s token for node %q: %w", i.name, workerName, err)
			}
			*i.token = t
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("create token transaction failed: %w", err))
		}
	}

	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.AddWorkerNode(ctx, tx, workerName)
	}); err != nil {
		return response.InternalError(fmt.Errorf("add worker node transaction failed: %w", err))
	}

	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteWorkerNodeToken(ctx, tx, workerName)
	}); err != nil {
		return response.InternalError(fmt.Errorf("delete worker node token transaction failed: %w", err))
	}

	return response.SyncResponse(true, &apiv1.WorkerNodeInfoResponse{
		CA:             cfg.Certificates.GetCACert(),
		APIServers:     servers,
		PodCIDR:        cfg.Network.GetPodCIDR(),
		ServiceCIDR:    cfg.Network.GetServiceCIDR(),
		KubeletToken:   kubeletToken,
		KubeProxyToken: proxyToken,
		ClusterDomain:  cfg.Kubelet.GetClusterDomain(),
		ClusterDNS:     cfg.Kubelet.GetClusterDNS(),
		CloudProvider:  cfg.Kubelet.GetCloudProvider(),
		KubeletCert:    workerCertificates.KubeletCert,
		KubeletKey:     workerCertificates.KubeletKey,
		K8sdPublicKey:  cfg.Certificates.GetK8sdPublicKey(),
	})
}
