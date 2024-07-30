package api

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
	v1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
)

func (e *Endpoints) postRefreshCertsPlan(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return refreshCertsPlanWorker(s, r, snap)
	} else {
		// TODO: Control Plane refresh
		return response.InternalError(fmt.Errorf("not implemented yet"))
	}

}

// refreshCertsPlanWorker generates the CSRs for the worker node and returns the seed and the names of the CSRs.
func refreshCertsPlanWorker(s *state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	client, err := snap.KubernetesNodeClient("")
	if err != nil {
		return response.InternalError(err)
	}

	var seed int32

	err = binary.Read(rand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return response.InternalError(err)
	}
	seed = seed & 0x7FFFFFFF

	log.Info("Generating CSRs for worker node")
	log.Info("Generating Kubelet Serving Certificate Signing Request")

	csrKubeletServing, pKeyKubeletServing, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", snap.Hostname()),
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		[]string{snap.Hostname()},
		[]net.IP{net.ParseIP(s.Address().Hostname())},
	)

	_, err = client.CreateCertificateSigningRequest(
		r.Context(),
		fmt.Sprintf("k8sd-%d-worker-kubelet-serving", seed),
		[]byte(csrKubeletServing),
		[]v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageServerAuth},
		[]string{"system:nodes"},
		"k8sd.io/kubelet-serving",
	)
	if err != nil {
		return response.InternalError(err)
	}

	log.Info("Generating Kubelet Client Certificate Signing Request")
	csrKubeletClient, pKeyKubeletClient, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", snap.Hostname()),
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
		nil,
	)

	_, err = client.CreateCertificateSigningRequest(
		r.Context(),
		fmt.Sprintf("k8sd-%d-worker-kubelet-client", seed),
		[]byte(csrKubeletClient),
		[]v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageClientAuth},
		nil,
		"k8sd.io/kubelet-client",
	)
	if err != nil {
		return response.InternalError(err)
	}

	log.Info("Generating Kube Proxy Client Certificate Signing Request")
	csrKubeProxy, pKeyKubeProxy, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName: "system:kube-proxy",
		},
		2048,
		nil,
		nil,
		nil,
	)

	_, err = client.CreateCertificateSigningRequest(
		r.Context(),
		fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", seed),
		[]byte(csrKubeProxy),
		[]v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageClientAuth},
		nil,
		"k8sd.io/kube-proxy-client",
	)

	result := apiv1.RefreshCertificatesPlanResponse{
		Seed: int(seed),
		CertificatesSigningRequests: []string{
			fmt.Sprintf("k8sd-%d-worker-kubelet-serving", seed),
			fmt.Sprintf("k8sd-%d-worker-kubelet-client", seed),
			fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", seed),
		},
	}

	log.Info("Writing new private keys")
	operations := []utils.FileOperations{
		{
			SourcePath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.tmp"),
			Content:     []byte(pKeyKubeletServing),
			Permissions: 0600,
		},
		{
			SourcePath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.tmp"),
			Content:     []byte(pKeyKubeletClient),
			Permissions: 0600,
		},
		{
			SourcePath:  filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.tmp"),
			Content:     []byte(pKeyKubeProxy),
			Permissions: 0600,
		},
	}

	err = utils.WriteFiles(operations)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, &result)

}

func (e *Endpoints) postRefreshCertsRun(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return refreshCertsRunWorker(r, snap)
	} else {
		// TODO: Control Plane refresh
		return response.InternalError(fmt.Errorf("not implemented yet"))
	}
}

// refreshCertsRunWorker refreshes the certificates for a worker node
func refreshCertsRunWorker(r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())
	req := apiv1.RefreshCertificatesRunRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	client, err := snap.KubernetesNodeClient("")
	if err != nil {
		return response.InternalError(err)
	}
	csrNames := []string{
		fmt.Sprintf("k8sd-%d-worker-kubelet-serving", req.Seed),
		fmt.Sprintf("k8sd-%d-worker-kubelet-client", req.Seed),
		fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", req.Seed),
	}

	certificates := pki.WorkerNodePKI{}

	operations := []utils.FileOperations{
		{
			SourcePath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet.crt"),
			BackupPath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet.crt.old"),
			Permissions: 0600,
		},
		{
			SourcePath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet.key"),
			BackupPath:  filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.old"),
			Permissions: 0600,
		},
		{
			SourcePath:  filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"),
			BackupPath:  filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf.old"),
			Permissions: 0600,
		},
		{
			SourcePath:  filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"),
			BackupPath:  filepath.Join(snap.KubernetesConfigDir(), "proxy.conf.old"),
			Permissions: 0600,
		},
	}

	log.Info("Backing up kubelet and kube-proxy certificates and configurations")
	utils.BackupFiles(operations)

	log.Info("Checking if the CSRs have been approved and issued")
	for _, csrName := range csrNames {
		csr, err := client.GetCertificateSigningRequest(r.Context(), csrName)
		if err != nil {
			return response.InternalError(err)
		}

		if !isCertificateSigningRequestApproved(csr) {
			log.Error(fmt.Errorf("CSR %s has not been approved", csrName), "CSR has not been approved")
			return response.InternalError(fmt.Errorf("CSR %s has not been approved", csrName))
		}

		if !isCertificateSigningRequestIssued(csr) {
			log.Error(fmt.Errorf("CSR %s has not been issued", csrName), "CSR has not been issued")
			return response.InternalError(fmt.Errorf("CSR %s has not been issued", csrName))
		}

		if _, _, err = pkiutil.LoadCertificate(string(csr.Status.Certificate), ""); err != nil {
			log.Error(err, fmt.Sprintf("failed to load certificate for CSR %s", csrName))
			return response.InternalError(fmt.Errorf("failed to load certificate for CSR %s: %w", csrName, err))
		}

		switch csrName {
		case fmt.Sprintf("k8sd-%d-worker-kubelet-serving", req.Seed):
			certificates.KubeletCert = string(csr.Status.Certificate)
		case fmt.Sprintf("k8sd-%d-worker-kubelet-client", req.Seed):
			certificates.KubeletClientCert = string(csr.Status.Certificate)
		case fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", req.Seed):
			certificates.KubeProxyClientCert = string(csr.Status.Certificate)
		default:
			log.Error(fmt.Errorf("unknown CSR %s", csrName), "Unknown CSR")
			return response.InternalError(fmt.Errorf("unknown CSR %s", csrName))
		}

	}

	// Read the CA and client CA
	ca, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "ca.crt"))
	if err != nil {
		return response.InternalError(err)
	}
	clientCA, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"))
	if err != nil {
		return response.InternalError(err)
	}

	// Read the new private keys
	bytesKubeletKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.tmp"))
	if err != nil {
		return response.InternalError(err)
	}

	bytesKubeletClientKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.tmp"))
	if err != nil {
		return response.InternalError(err)
	}

	bytesKubeProxyKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.tmp"))
	if err != nil {
		return response.InternalError(err)
	}

	certificates.CACert = string(ca)
	certificates.ClientCACert = string(clientCA)
	certificates.KubeletKey = string(bytesKubeletKey)
	certificates.KubeletClientKey = string(bytesKubeletClientKey)
	certificates.KubeProxyClientKey = string(bytesKubeProxyKey)

	log.Info("Ensuring worker PKI")
	if _, err = setup.EnsureWorkerPKI(snap, &certificates); err != nil {
		return response.InternalError(err)
	}

	// Kubeconfigs
	log.Info("Generating new kubeconfigs")
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), "127.0.0.1:6443", certificates.CACert, certificates.KubeletClientCert, certificates.KubeletClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kubelet kubeconfig: %w", err))
	}
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), "127.0.0.1:6443", certificates.CACert, certificates.KubeProxyClientCert, certificates.KubeProxyClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err))
	}

	// Restart the services
	log.Info("Restarting kubelet and kube-proxy")
	if err := snap.RestartService(r.Context(), "kubelet"); err != nil {
		return response.InternalError(err)
	}
	if err := snap.RestartService(r.Context(), "kube-proxy"); err != nil {
		return response.InternalError(err)
	}

	// Remove the new private keys
	log.Info("Removing temporal private keys")
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.tmp")); err != nil {
		return response.InternalError(err)
	}
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.tmp")); err != nil {
		return response.InternalError(err)
	}
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.tmp")); err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, nil)

}

// isCertificateSigningRequestApproved checks if the certificate signing request is approved.
func isCertificateSigningRequestApproved(csr *v1.CertificateSigningRequest) bool {
	for _, condition := range csr.Status.Conditions {
		if condition.Type == v1.CertificateApproved && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// isCertificateSigningRequestIssued checks if the certificate signing request is issued.
func isCertificateSigningRequestIssued(csr *v1.CertificateSigningRequest) bool {
	return len(csr.Status.Certificate) > 0
}
