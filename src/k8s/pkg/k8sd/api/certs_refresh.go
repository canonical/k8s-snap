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
		return refreshCertsPlanWorker(s, snap)

	} else {
		// TODO: Control Plane refresh
		return response.InternalError(fmt.Errorf("not implemented yet"))
	}

}

// refreshCertsPlanWorker generates the CSRs for the worker node and returns the seed and the names of the CSRs.
func refreshCertsPlanWorker(s *state.State, snap snap.Snap) response.Response {
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
		s.Context,
		fmt.Sprintf("k8sd-%d-worker-kubelet-serving", seed),
		[]byte(csrKubeletServing),
		[]v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageServerAuth},
		[]string{"system:nodes"},
		"k8sd.io/kubelet-serving",
	)
	if err != nil {
		return response.InternalError(err)
	}

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
		s.Context,
		fmt.Sprintf("k8sd-%d-worker-kubelet-client", seed),
		[]byte(csrKubeletClient),
		[]v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageClientAuth},
		nil,
		"k8sd.io/kubelet-client",
	)
	if err != nil {
		return response.InternalError(err)
	}

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
		s.Context,
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

	err = os.WriteFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.new"), []byte(pKeyKubeletServing), 0600)
	if err != nil {
		return response.InternalError(err)
	}
	err = os.WriteFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.new"), []byte(pKeyKubeletClient), 0600)
	if err != nil {
		return response.InternalError(err)
	}
	err = os.WriteFile(filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.new"), []byte(pKeyKubeProxy), 0600)
	if err != nil {
		return response.InternalError(err)
	}
	return response.SyncResponse(true, &result)

}

func (e *Endpoints) postRefreshCertsRun(s *state.State, r *http.Request) response.Response {
	// TODO: Control Plane refresh
	req := apiv1.RefreshCertificatesRunRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}
	snap := e.provider.Snap()

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

	for _, csrName := range csrNames {
		csr, err := client.GetCertificateSigningRequest(s.Context, csrName)
		if err != nil {
			return response.InternalError(err)
		}

		if !isCertificateApproved(csr.Status.Conditions) {
			return response.InternalError(fmt.Errorf("CSR %s is not issued", csrName))
		}

		if len(csr.Status.Certificate) == 0 {
			return response.InternalError(fmt.Errorf("CSR %s is missing certificate", csrName))
		}

		_, _, err = pkiutil.LoadCertificate(string(csr.Status.Certificate), "")
		if err != nil {
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
	bytesKubeletKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.new"))
	if err != nil {
		return response.InternalError(err)
	}

	bytesKubeletClientKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.new"))
	if err != nil {
		return response.InternalError(err)
	}

	bytesKubeProxyKey, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.new"))
	if err != nil {
		return response.InternalError(err)
	}

	certificates.CACert = string(ca)
	certificates.ClientCACert = string(clientCA)
	certificates.KubeletKey = string(bytesKubeletKey)
	certificates.KubeletClientKey = string(bytesKubeletClientKey)
	certificates.KubeProxyClientKey = string(bytesKubeProxyKey)

	_, err = setup.EnsureWorkerPKI(snap, &certificates)
	if err != nil {
		return response.InternalError(err)
	}

	// Kubeconfigs
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), "127.0.0.1:6443", certificates.CACert, certificates.KubeletClientCert, certificates.KubeletClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kubelet kubeconfig: %w", err))
	}
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), "127.0.0.1:6443", certificates.CACert, certificates.KubeProxyClientCert, certificates.KubeProxyClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err))
	}

	// Restart the services
	if err := snap.RestartService(s.Context, "kubelet"); err != nil {
		return response.InternalError(err)
	}
	if err := snap.RestartService(s.Context, "kube-proxy"); err != nil {
		return response.InternalError(err)
	}

	// Remove the new private keys
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kubelet.key.new")); err != nil {
		return response.InternalError(err)
	}
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kubelet-client.key.new")); err != nil {
		return response.InternalError(err)
	}
	if err := os.Remove(filepath.Join(snap.KubernetesPKIDir(), "kube-proxy.key.new")); err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, nil)

}

// isCertificateApproved checks if the certificate signing request is approved.
func isCertificateApproved(conditions []v1.CertificateSigningRequestCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.CertificateApproved && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
