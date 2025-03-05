package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	nodeutil "github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) postRefreshCertsUpdate(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return refreshCertsUpdateWorker(s, r, snap)
	}
	return refreshCertsUpdateControlPlane(s, r, snap)
}

// refreshCertsUpdateControlPlane updates the external certificates for a control plane node.
func refreshCertsUpdateControlPlane(s state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	req := apiv1.RefreshCertificatesUpdateRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	clusterConfig, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to recover cluster config: %w", err))
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return response.InternalError(fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname()))
	}

	var localhostAddress string
	if nodeIP.To4() == nil {
		localhostAddress = "[::1]"
	} else {
		localhostAddress = "127.0.0.1"
	}

	serviceIPs, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(clusterConfig.Network.GetServiceCIDR())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get IP address(es) from ServiceCIDR %q: %w", clusterConfig.Network.GetServiceCIDR(), err))
	}

	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:  s.Name(),
		IPSANs:    append([]net.IP{nodeIP}, serviceIPs...),
		NotBefore: time.Now(),
	})
	certificates.CACert = clusterConfig.Certificates.GetCACert()
	certificates.CAKey = clusterConfig.Certificates.GetCAKey()
	certificates.ClientCACert = clusterConfig.Certificates.GetClientCACert()
	certificates.ClientCAKey = clusterConfig.Certificates.GetClientCAKey()
	certificates.FrontProxyCACert = clusterConfig.Certificates.GetFrontProxyCACert()
	certificates.FrontProxyCAKey = clusterConfig.Certificates.GetFrontProxyCAKey()
	certificates.K8sdPrivateKey = clusterConfig.Certificates.GetK8sdPrivateKey()
	certificates.K8sdPublicKey = clusterConfig.Certificates.GetK8sdPublicKey()
	certificates.ServiceAccountKey = clusterConfig.Certificates.GetServiceAccountKey()

	certificates.AdminClientCert = req.GetAdminClientCert()
	certificates.AdminClientKey = req.GetAdminClientKey()
	certificates.APIServerKubeletClientCert = req.GetAPIServerKubeletClientCert()
	certificates.APIServerKubeletClientKey = req.GetAPIServerKubeletClientKey()
	certificates.KubeControllerManagerClientCert = req.GetKubeControllerManagerClientCert()
	certificates.KubeControllerManagerClientKey = req.GetKubeControllerManagerClientKey()
	certificates.KubeSchedulerClientCert = req.GetKubeSchedulerClientCert()
	certificates.KubeSchedulerClientKey = req.GetKubeSchedulerClientKey()
	certificates.APIServerCert = req.GetAPIServerCert()
	certificates.APIServerKey = req.GetAPIServerKey()
	certificates.KubeProxyClientCert = req.GetKubeProxyClientCert()
	certificates.KubeProxyClientKey = req.GetKubeProxyClientKey()
	certificates.KubeletClientCert = req.GetKubeletClientCert()
	certificates.KubeletClientKey = req.GetKubeletClientKey()
	certificates.KubeletCert = req.GetKubeletCert()
	certificates.KubeletKey = req.GetKubeletKey()
	certificates.FrontProxyClientCert = req.GetFrontProxyClientCert()
	certificates.FrontProxyClientKey = req.GetFrontProxyClientKey()

	if err := certificates.CompleteCertificates(); err != nil {
		return response.InternalError(fmt.Errorf("failed to verify certificates: %w", err))
	}

	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to write control plane certificates: %w", err))
	}

	if err := setup.SetupControlPlaneKubeconfigs(snap.KubernetesConfigDir(), localhostAddress, clusterConfig.APIServer.GetSecurePort(), *certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate control plane kubeconfigs: %w", err))
	}

	restartFn := func(ctx context.Context) error {
		if err := snaputil.RestartControlPlaneServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to restart control plane services: %w", err)
		}
		return nil
	}
	readyCh := nodeutil.StartAsyncRestart(log, restartFn)

	return utils.SyncManualResponseWithSignal(r, readyCh, apiv1.RefreshCertificatesUpdateResponse{})
}

// refreshCertsUpdateWorker updates the external certificates for a worker node.
func refreshCertsUpdateWorker(s state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	req := apiv1.RefreshCertificatesUpdateRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	clusterConfig, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to recover cluster config: %w", err))
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return response.InternalError(fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname()))
	}

	var localhostAddress string
	if nodeIP.To4() == nil {
		localhostAddress = "[::1]"
	} else {
		localhostAddress = "127.0.0.1"
	}

	var certificates pki.WorkerNodePKI
	certificates.CACert = clusterConfig.Certificates.GetCACert()
	certificates.ClientCACert = clusterConfig.Certificates.GetClientCACert()

	certificates.KubeletCert = req.GetKubeletCert()
	certificates.KubeletKey = req.GetKubeletKey()
	certificates.KubeletClientCert = req.GetKubeletClientCert()
	certificates.KubeletClientKey = req.GetKubeletClientKey()
	certificates.KubeProxyClientCert = req.GetKubeProxyClientCert()
	certificates.KubeProxyClientKey = req.GetKubeProxyClientKey()

	if err := certificates.CompleteCertificates(); err != nil {
		return response.InternalError(fmt.Errorf("failed to verify certificates: %w", err))
	}

	if _, err := setup.EnsureWorkerPKI(snap, &certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to write worker certificates: %w", err))
	}

	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), fmt.Sprintf("%s:%d", localhostAddress, clusterConfig.APIServer.GetSecurePort()), certificates.CACert, certificates.KubeletClientCert, certificates.KubeletClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to write kubeconfig %s: %w", filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), err))
	}

	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), fmt.Sprintf("%s:%d", localhostAddress, clusterConfig.APIServer.GetSecurePort()), certificates.CACert, certificates.KubeProxyClientCert, certificates.KubeProxyClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to write kubeconfig %s: %w", filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), err))
	}

	restartFn := func(ctx context.Context) error {
		if err := snap.RestartService(ctx, "kubelet"); err != nil {
			return fmt.Errorf("failed to restart kubelet: %w", err)
		}

		if err := snap.RestartService(ctx, "kube-proxy"); err != nil {
			return fmt.Errorf("failed to restart kube-proxy: %w", err)
		}
		return nil
	}

	readyCh := nodeutil.StartAsyncRestart(log, restartFn)

	return utils.SyncManualResponseWithSignal(r, readyCh, apiv1.RefreshCertificatesUpdateResponse{})
}
