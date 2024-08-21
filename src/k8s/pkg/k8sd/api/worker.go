package api

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) postWorkerInfo(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.GetWorkerJoinInfoRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	// Existence of this header is already checked in the access handler.
	workerName := r.Header.Get("worker-name")
	nodeIP := net.ParseIP(req.Address)
	if nodeIP == nil {
		return response.BadRequest(fmt.Errorf("failed to parse node IP address %s", req.Address))
	}

	cfg, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	// NOTE: Set the notBefore certificate time to the current time.
	notBefore := time.Now()
	// NOTE: Default certificate expiration is set to 10 years.
	defaultExpiration := notBefore.AddDate(10, 0, 0)

	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{NotBefore: notBefore, NotAfter: defaultExpiration})
	certificates.CACert = cfg.Certificates.GetCACert()
	certificates.CAKey = cfg.Certificates.GetCAKey()
	certificates.ClientCACert = cfg.Certificates.GetClientCACert()
	certificates.ClientCAKey = cfg.Certificates.GetClientCAKey()
	workerCertificates, err := certificates.CompleteWorkerNodePKI(workerName, nodeIP, 2048)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to generate worker PKI: %w", err))
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create kubernetes client: %w", err))
	}
	if err := client.WaitKubernetesEndpointAvailable(r.Context()); err != nil {
		return response.InternalError(fmt.Errorf("kubernetes endpoints not ready yet: %w", err))
	}
	servers, err := client.GetKubeAPIServerEndpoints(r.Context())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve list of known kube-apiserver endpoints: %w", err))
	}

	workerToken := r.Header.Get("worker-token")
	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteWorkerNodeToken(ctx, tx, workerToken)
	}); err != nil {
		return response.InternalError(fmt.Errorf("delete worker node token transaction failed: %w", err))
	}

	return response.SyncResponse(true, &apiv1.GetWorkerJoinInfoResponse{
		CACert:              cfg.Certificates.GetCACert(),
		ClientCACert:        cfg.Certificates.GetClientCACert(),
		APIServers:          servers,
		PodCIDR:             cfg.Network.GetPodCIDR(),
		ServiceCIDR:         cfg.Network.GetServiceCIDR(),
		ClusterDomain:       cfg.Kubelet.GetClusterDomain(),
		ClusterDNS:          cfg.Kubelet.GetClusterDNS(),
		CloudProvider:       cfg.Kubelet.GetCloudProvider(),
		KubeletCert:         workerCertificates.KubeletCert,
		KubeletKey:          workerCertificates.KubeletKey,
		KubeletClientCert:   workerCertificates.KubeletClientCert,
		KubeletClientKey:    workerCertificates.KubeletClientKey,
		KubeProxyClientCert: workerCertificates.KubeProxyClientCert,
		KubeProxyClientKey:  workerCertificates.KubeProxyClientKey,
		K8sdPublicKey:       cfg.Certificates.GetK8sdPublicKey(),
	})
}
