package api

import (
	"crypto/x509/pkix"
	"fmt"
	"math"
	"math/rand"
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
	"golang.org/x/sync/errgroup"
	certificatesv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Endpoints) postRefreshCertsPlan(s *state.State, r *http.Request) response.Response {
	log := log.FromContext(r.Context())

	log.Info("Generating random seed for certificates refresh")
	seed := rand.Intn(math.MaxInt)

	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return response.SyncResponse(true, apiv1.RefreshCertificatesPlanResponse{
			Seed: seed,
			CertificatesSigningRequests: []string{
				fmt.Sprintf("k8sd-%d-worker-kubelet-serving", seed),
				fmt.Sprintf("k8sd-%d-worker-kubelet-client", seed),
				fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", seed),
			},
		},
		)
	}

	return response.SyncResponse(true, apiv1.RefreshCertificatesPlanResponse{
		Seed: seed,
	})

}

func (e *Endpoints) postRefreshCertsRun(s *state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return refreshCertsRunWorker(s, r, snap)
	}
	// TODO: Control Plane refresh
	return response.InternalError(fmt.Errorf("not implemented yet"))
}

// refreshCertsRunWorker refreshes the certificates for a worker node
func refreshCertsRunWorker(s *state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	req := apiv1.RefreshCertificatesRunRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	client, err := snap.KubernetesNodeClient("")
	if err != nil {
		return response.InternalError(err)
	}

	log.Info("Generating CSRs for worker node")
	certificates := pki.WorkerNodePKI{}
	// Read the CA and client CA
	ca, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "ca.crt"))
	if err != nil {
		return response.InternalError(err)
	}
	clientCA, err := os.ReadFile(filepath.Join(snap.KubernetesPKIDir(), "client-ca.crt"))
	if err != nil {
		return response.InternalError(err)
	}

	certificates.CACert = string(ca)
	certificates.ClientCACert = string(clientCA)

	g, errGroupCTX := errgroup.WithContext(r.Context())

	for _, csr := range []struct {
		name         string
		commonName   string
		organization []string
		usages       []v1.KeyUsage
		hostnames    []string
		ips          []net.IP
		signerName   string
	}{
		{
			name:         fmt.Sprintf("k8sd-%d-worker-kubelet-serving", req.Seed),
			commonName:   fmt.Sprintf("system:node:%s", snap.Hostname()),
			organization: []string{"system:nodes"},
			usages:       []v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageServerAuth},
			hostnames:    []string{snap.Hostname()},
			ips:          []net.IP{net.ParseIP(s.Address().Hostname())},
			signerName:   "k8sd.io/kubelet-serving",
		},
		{
			name:         fmt.Sprintf("k8sd-%d-worker-kubelet-client", req.Seed),
			commonName:   fmt.Sprintf("system:node:%s", snap.Hostname()),
			organization: []string{"system:nodes"},
			usages:       []v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageClientAuth},
			signerName:   "k8sd.io/kubelet-client",
		},
		{
			name:       fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", req.Seed),
			commonName: "system:kube-proxy",
			usages:     []v1.KeyUsage{v1.UsageDigitalSignature, v1.UsageKeyEncipherment, v1.UsageClientAuth},
			signerName: "k8sd.io/kube-proxy-client",
		},
	} {
		csr := csr
		g.Go(func() error {
			componentCSR, pKey, err := pkiutil.GenerateCSR(
				pkix.Name{
					CommonName:   csr.commonName,
					Organization: csr.organization,
				},
				2048,
				nil,
				csr.hostnames,
				csr.ips,
			)
			if err != nil {
				return fmt.Errorf("failed to generate CSR for %s: %w", csr.name, err)
			}

			_, err = client.CertificatesV1().CertificateSigningRequests().Create(errGroupCTX, &certificatesv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: csr.name,
				},
				Spec: certificatesv1.CertificateSigningRequestSpec{
					Request:    []byte(componentCSR),
					Usages:     csr.usages,
					SignerName: csr.signerName,
				},
			}, metav1.CreateOptions{})

			if err != nil {
				return fmt.Errorf("failed to create CSR for %s: %w", csr.name, err)
			}

			for {
				select {
				case <-errGroupCTX.Done():
					return nil
				default:
					request, err := client.CertificatesV1().CertificateSigningRequests().Get(errGroupCTX, csr.name, metav1.GetOptions{})
					if err != nil {
						log.Error(err, fmt.Sprintf("failed to get CSR %s", csr.name))
						continue
					}

					approved, err := isCertificateSigningRequestApproved(request)
					if err != nil {
						log.Error(err, fmt.Sprintf("failed to check approval status for CSR %s", csr.name))
						return fmt.Errorf("failed to check approval status for CSR %s: %w", csr.name, err)
					}

					if approved && isCertificateSigningRequestIssued(request) {
						if _, _, err = pkiutil.LoadCertificate(string(request.Status.Certificate), ""); err != nil {
							log.Error(err, fmt.Sprintf("failed to load certificate for CSR %s", csr.name))
						}
						switch csr.name {
						case fmt.Sprintf("k8sd-%d-worker-kubelet-serving", req.Seed):
							certificates.KubeletCert = string(request.Status.Certificate)
							certificates.KubeletKey = string(pKey)
						case fmt.Sprintf("k8sd-%d-worker-kubelet-client", req.Seed):
							certificates.KubeletClientCert = string(request.Status.Certificate)
							certificates.KubeletClientKey = string(pKey)
						case fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", req.Seed):
							certificates.KubeProxyClientCert = string(request.Status.Certificate)
							certificates.KubeProxyClientKey = string(pKey)
						}
						return nil
					}
				}
			}
		})

	}

	if err := g.Wait(); err != nil {
		return response.InternalError(err)
	}

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

	return response.SyncResponse(true, nil)

}

// isCertificateSigningRequestApproved checks if the certificate signing request is approved.
// It returns true if the CSR is approved, false if it is pending, and an error if it is denied.
func isCertificateSigningRequestApproved(csr *v1.CertificateSigningRequest) (bool, error) {
	for _, condition := range csr.Status.Conditions {
		if condition.Type == v1.CertificateApproved && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
		if condition.Type == v1.CertificateDenied && condition.Status == corev1.ConditionTrue {
			return false, fmt.Errorf("CSR %s was denied: %s", csr.Name, condition.Reason)
		}
	}
	return false, nil
}

// isCertificateSigningRequestIssued checks if the certificate signing request is issued.
func isCertificateSigningRequestIssued(csr *v1.CertificateSigningRequest) bool {
	return len(csr.Status.Certificate) > 0
}
