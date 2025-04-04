package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"net"
	"net/http"
	"path/filepath"
	"slices"
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
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
	"golang.org/x/sync/errgroup"
	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertFrontProxyClient        = "front-proxy-client"
	CertAPIServerKubeletClient  = "apiserver-kubelet-client"
	CertAPIServer               = "apiserver"
	CertKubelet                 = "kubelet"
	CertAdminClient             = "admin.conf"
	CertSchedulerClient         = "scheduler.conf"
	CertControllerManagerClient = "controller.conf"
	CertKubeletClient           = "kubelet.conf"
	CertProxyClient             = "proxy.conf"
)

// controlPlaneCertificateMarker manages pointers to control plane PKI fields.
type controlPlaneCertificateMarker struct {
	certificates map[string]*string
	keys         map[string]*string
}

// csrDefinition holds the paramaters needed to generate and request a specific CSR.
type csrDefinition struct {
	CertName     string
	CSRBaseName  string
	CommonName   string
	Organization []string
	Usages       []certv1.KeyUsage
	SignerName   string
	NeedsSANs    bool

	targetCert *string
	targetKey  *string
}

// Map roles to the set of valid certificate names for that role.
var certsByRole = map[apiv1.ClusterRole]map[string]struct{}{
	apiv1.ClusterRoleWorker: {
		CertKubelet:       {},
		CertKubeletClient: {},
		CertProxyClient:   {},
	},
	apiv1.ClusterRoleControlPlane: {
		CertAdminClient:             {},
		CertFrontProxyClient:        {},
		CertAPIServerKubeletClient:  {},
		CertSchedulerClient:         {},
		CertControllerManagerClient: {},
		CertAPIServer:               {},
		CertKubeletClient:           {},
		CertKubelet:                 {},
		CertProxyClient:             {},
	},
}

// newControlPlaneCertificateMarker initializes the marker with pointers to
// the PKI struct fields.
func newControlPlaneCertificateMarker(pki *pki.ControlPlanePKI) *controlPlaneCertificateMarker {
	marker := &controlPlaneCertificateMarker{}
	marker.certificates = make(map[string]*string, 9)
	marker.keys = make(map[string]*string, 9)

	marker.certificates[CertAdminClient] = &pki.AdminClientCert
	marker.keys[CertAdminClient] = &pki.AdminClientKey

	marker.certificates[CertFrontProxyClient] = &pki.FrontProxyClientCert
	marker.keys[CertFrontProxyClient] = &pki.FrontProxyClientKey

	marker.certificates[CertAPIServerKubeletClient] = &pki.APIServerKubeletClientCert
	marker.keys[CertAPIServerKubeletClient] = &pki.APIServerKubeletClientKey

	marker.certificates[CertSchedulerClient] = &pki.KubeSchedulerClientCert
	marker.keys[CertSchedulerClient] = &pki.KubeSchedulerClientKey

	marker.certificates[CertControllerManagerClient] = &pki.KubeControllerManagerClientCert
	marker.keys[CertControllerManagerClient] = &pki.KubeControllerManagerClientKey

	marker.certificates[CertAPIServer] = &pki.APIServerCert
	marker.keys[CertAPIServer] = &pki.APIServerKey

	marker.certificates[CertKubeletClient] = &pki.KubeletClientCert
	marker.keys[CertKubeletClient] = &pki.KubeletClientKey

	marker.certificates[CertKubelet] = &pki.KubeletCert
	marker.keys[CertKubelet] = &pki.KubeletKey

	marker.certificates[CertProxyClient] = &pki.KubeProxyClientCert
	marker.keys[CertProxyClient] = &pki.KubeProxyClientKey

	return marker
}

// markCertificatesForRefresh clears the string values pointed to for the
// specified certs.
func (c *controlPlaneCertificateMarker) markCertificatesForRefresh(certificates []string) error {
	for _, cert := range certificates {
		certValuePtr, found := c.certificates[cert]
		if !found {
			return fmt.Errorf("invalid certificate name for control plane: %s", cert)
		}
		*certValuePtr = ""
		keyValuePtr, found := c.keys[cert]
		if !found {
			return fmt.Errorf("invalid key name for control plane: %s", cert)
		}
		*keyValuePtr = ""
	}
	return nil
}

// validateCertificatesByRole checks if the requested certs are valid for the
// given role.
func validateCertificatesByRole(role apiv1.ClusterRole, certs []string) error {
	validCerts, ok := certsByRole[role]
	if !ok {
		return fmt.Errorf("unknown role: %s", role)
	}

	for _, cert := range certs {
		if _, found := validCerts[cert]; !found {
			return fmt.Errorf("invalid certificates for %s node: %s", role, cert)
		}
	}
	return nil
}

func (e *Endpoints) postRefreshCertsPlan(s state.State, r *http.Request) response.Response {
	req := apiv1.RefreshCertificatesPlanRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	seedBigInt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to generate seed: %w", err))
	}
	seed := int(seedBigInt.Int64())

	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}

	var role apiv1.ClusterRole
	var certsToPlan []string
	if isWorker {
		role = apiv1.ClusterRoleWorker
		if len(req.Certificates) > 0 {
			certsToPlan = req.Certificates
		} else {
			certsToPlan = getAllCertsForRole(apiv1.ClusterRoleWorker)
		}
	} else {
		role = apiv1.ClusterRoleControlPlane
		if len(req.Certificates) > 0 {
			certsToPlan = req.Certificates
		}
	}

	if len(req.Certificates) > 0 {
		if err := validateCertificatesByRole(role, certsToPlan); err != nil {
			return response.BadRequest(fmt.Errorf("failed to validate requested certificates: %w", err))
		}
	}

	resp := apiv1.RefreshCertificatesPlanResponse{
		Seed: seed,
	}

	if isWorker {
		workerCSRDefs := getWorkerCSRDefinitions(snap.Hostname())
		resp.CertificateSigningRequests = make([]string, 0, len(certsToPlan))

		for _, certName := range certsToPlan {
			csrDef, found := workerCSRDefs[certName]
			if !found {
				// NOTE (mateoflorido): This should ideally not happen due to prior validation.
				return response.BadRequest(fmt.Errorf("CSR definition not found for validated certificate %q", certName))
			}
			csrObjectName := fmt.Sprintf("k8sd-%d-%s", seed, csrDef.CSRBaseName)
			resp.CertificateSigningRequests = append(resp.CertificateSigningRequests, csrObjectName)
		}
	}

	return response.SyncResponse(true, resp)
}

func (e *Endpoints) postRefreshCertsRun(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return refreshCertsRunWorker(s, r, snap)
	}
	return refreshCertsRunControlPlane(s, r, snap)
}

// refreshCertsRunControlPlane refreshes the certificates for a control plane node.
func refreshCertsRunControlPlane(s state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	req := apiv1.RefreshCertificatesRunRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	certsToRefresh := req.Certificates
	if len(certsToRefresh) > 0 {
		if err := validateCertificatesByRole(apiv1.ClusterRoleControlPlane, certsToRefresh); err != nil {
			return response.BadRequest(fmt.Errorf("failed to validate requested certificates: %w", err))
		}
	} else {
		certsToRefresh = getAllCertsForRole(apiv1.ClusterRoleControlPlane)
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

	extraIPs, extraNames := utils.SplitIPAndDNSSANs(req.ExtraSANs)

	// NOTE: Set the notBefore certificate time to the current time.
	notBefore := time.Now()

	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append(append([]net.IP{nodeIP}, serviceIPs...), extraIPs...),
		NotBefore:                 notBefore,
		NotAfter:                  utils.SecondsToExpirationDate(notBefore, req.ExpirationSeconds),
		DNSSANs:                   extraNames,
		AllowSelfSignedCA:         true,
		IncludeMachineAddressSANs: true,
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

	manager := newControlPlaneCertificateMarker(certificates)
	if err := setup.ReadControlPlanePKI(snap, certificates, true); err != nil {
		return response.InternalError(fmt.Errorf("failed to read managed control plane certificates: %w", err))
	}

	if err := setup.ReadControlPlanePKI(snap, certificates, false); err != nil {
		return response.InternalError(fmt.Errorf("failed to read unmanaged control plane certificates: %w", err))
	}

	if err := manager.markCertificatesForRefresh(certsToRefresh); err != nil {
		return response.InternalError(fmt.Errorf("failed to mark certificates for refresh: %w", err))
	}

	if err := certificates.CompleteCertificates(); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate new control plane certificates: %w", err))
	}

	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to write control plane certificates: %w", err))
	}

	if err := setup.SetupControlPlaneKubeconfigs(snap.KubernetesConfigDir(), localhostAddress, clusterConfig.APIServer.GetSecurePort(), *certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate control plane kubeconfigs: %w", err))
	}

	// NOTE: Restart the control plane services in a separate goroutine to avoid
	// restarting the API server, which would break the k8sd proxy connection
	// and cause missed responses in the proxy side.
	restartFn := func(ctx context.Context) error {
		if err := snaputil.RestartControlPlaneServices(ctx, snap); err != nil {
			return fmt.Errorf("failed to restart control plane services: %w", err)
		}
		return nil
	}
	readyCh := nodeutil.StartAsyncRestart(log, restartFn)

	apiServerCert, _, err := pkiutil.LoadCertificate(certificates.APIServerCert, "")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read kubelet certificate: %w", err))
	}

	expirationTimeUNIX := apiServerCert.NotAfter.Unix()

	return utils.SyncManualResponseWithSignal(r, readyCh, apiv1.RefreshCertificatesRunResponse{
		ExpirationSeconds: int(expirationTimeUNIX),
	})
}

// refreshCertsRunWorker refreshes the certificates for a worker node.
func refreshCertsRunWorker(s state.State, r *http.Request, snap snap.Snap) response.Response {
	log := log.FromContext(r.Context())

	req := apiv1.RefreshCertificatesRunRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	certsToRefresh := req.Certificates
	if len(certsToRefresh) == 0 {
		certsToRefresh = getAllCertsForRole(apiv1.ClusterRoleWorker)
	} else {
		if err := validateCertificatesByRole(apiv1.ClusterRoleWorker, certsToRefresh); err != nil {
			return response.BadRequest(fmt.Errorf("failed to validate requested certificates: %w", err))
		}
	}

	client, err := snap.KubernetesNodeClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get Kubernetes client: %w", err))
	}

	var certificates pki.WorkerNodePKI

	clusterConfig, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster configuration: %w", err))
	}

	if clusterConfig.Certificates.CACert == nil || clusterConfig.Certificates.ClientCACert == nil {
		return response.InternalError(fmt.Errorf("missing CA certificates from cluster config"))
	}

	certificates.CACert = clusterConfig.Certificates.GetCACert()
	certificates.ClientCACert = clusterConfig.Certificates.GetClientCACert()

	k8sdPublicKey, err := pkiutil.LoadRSAPublicKey(clusterConfig.Certificates.GetK8sdPublicKey())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to load k8sd public key: %w", err))
	}

	hostnames := []string{snap.Hostname()}
	ips := []net.IP{net.ParseIP(s.Address().Hostname())}

	extraIPs, extraNames := utils.SplitIPAndDNSSANs(req.ExtraSANs)
	hostnames = append(hostnames, extraNames...)
	ips = append(ips, extraIPs...)

	workerCSRDefs := getWorkerCSRDefinitions(snap.Hostname())
	g, ctx := errgroup.WithContext(r.Context())
	expirationSeconds := int32(req.ExpirationSeconds)

	for _, certName := range certsToRefresh {
		csrDef, found := workerCSRDefs[certName]
		if !found {
			// NOTE (mateoflorido): Should not happen after validation.
			return response.InternalError(fmt.Errorf("CSR definition not found for %q", certName))
		}
		switch certName {
		case CertKubelet:
			csrDef.targetCert = &certificates.KubeletCert
			csrDef.targetKey = &certificates.KubeletKey
		case CertKubeletClient:
			csrDef.targetCert = &certificates.KubeletClientCert
			csrDef.targetKey = &certificates.KubeletClientKey
		case CertProxyClient:
			csrDef.targetCert = &certificates.KubeProxyClientCert
			csrDef.targetKey = &certificates.KubeProxyClientKey
		default:
			// NOTE (mateoflorido): Should not happen.
			return response.InternalError(fmt.Errorf("unhandled certificate name %q for worker CSR", certName))
		}

		localCSRDef := csrDef

		g.Go(func() error {
			csrObjectName := fmt.Sprintf("k8sd-%d-%s", req.Seed, localCSRDef.CSRBaseName)

			var csrHostnames []string
			var csrIPs []net.IP
			if localCSRDef.NeedsSANs {
				csrHostnames = hostnames
				csrIPs = ips
			}

			csrPEM, keyPEM, err := pkiutil.GenerateCSR(
				pkix.Name{
					CommonName:   localCSRDef.CommonName,
					Organization: localCSRDef.Organization,
				},
				2048,
				csrHostnames,
				csrIPs,
			)
			if err != nil {
				return fmt.Errorf("failed to generate CSR for %s: %w", csrObjectName, err)
			}

			// Obtain the SHA256 sum of the CSR request.
			hash := sha256.New()
			_, err = hash.Write([]byte(csrPEM))
			if err != nil {
				return fmt.Errorf("failed to checksum CSR %s, err: %w", csrObjectName, err)
			}
			signature, err := rsa.EncryptPKCS1v15(rand.Reader, k8sdPublicKey, hash.Sum(nil))
			if err != nil {
				return fmt.Errorf("failed to sign CSR %s, err: %w", csrObjectName, err)
			}
			signatureB64 := base64.StdEncoding.EncodeToString(signature)

			k8sCSR := &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: csrObjectName,
					Annotations: map[string]string{
						"k8sd.io/signature": signatureB64,
						"k8sd.io/node":      snap.Hostname(),
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:           []byte(csrPEM),
					ExpirationSeconds: &expirationSeconds,
					Usages:            localCSRDef.Usages,
					SignerName:        localCSRDef.SignerName,
				},
			}

			if _, err = client.CertificatesV1().CertificateSigningRequests().Create(ctx, k8sCSR, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("failed to create CSR for %s: %w", csrObjectName, err)
			}

			if err := client.WatchCertificateSigningRequest(
				ctx,
				csrObjectName,
				func(request *certv1.CertificateSigningRequest) (bool, error) {
					return verifyCSRAndSetPKI(request, keyPEM, localCSRDef.targetCert, csrDef.targetKey)
				},
			); err != nil {
				log.Error(err, "Failed to watch CSR %s: %v", csrObjectName, err)
				return fmt.Errorf("certificate signing request %s failed: %w", csrObjectName, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return response.InternalError(fmt.Errorf("failed to get one or more worker node certificates: %w", err))
	}

	if _, err = setup.EnsureWorkerPKI(snap, &certificates); err != nil {
		return response.InternalError(fmt.Errorf("failed to write worker PKI: %w", err))
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

	// Kubeconfigs
	apiServerEndpoint := fmt.Sprintf("%s:%d", localhostAddress, clusterConfig.APIServer.GetSecurePort())
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"), apiServerEndpoint, certificates.CACert, certificates.KubeletClientCert, certificates.KubeletClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kubelet kubeconfig: %w", err))
	}
	if err := setup.Kubeconfig(filepath.Join(snap.KubernetesConfigDir(), "proxy.conf"), apiServerEndpoint, certificates.CACert, certificates.KubeProxyClientCert, certificates.KubeProxyClientKey); err != nil {
		return response.InternalError(fmt.Errorf("failed to generate kube-proxy kubeconfig: %w", err))
	}

	// NOTE: Restart the worker services in a separate goroutine to avoid
	// restarting the kube-proxy and kubelet, which would break the
	// proxy connection and cause missed responses in the proxy side.
	restartFn := func(ctx context.Context) error {
		restartKubelet := slices.Contains(certsToRefresh, CertKubelet) || slices.Contains(certsToRefresh, CertKubeletClient)
		restartProxy := slices.Contains(certsToRefresh, CertProxyClient)

		if restartKubelet {
			if err := snap.RestartServices(ctx, []string{"kubelet"}); err != nil {
				return fmt.Errorf("failed to restart kubelet: %w", err)
			}
		}

		if restartProxy {
			if err := snap.RestartServices(ctx, []string{"kube-proxy"}); err != nil {
				return fmt.Errorf("failed to restart kube-proxy: %w", err)
			}
		}
		return nil
	}
	readyCh := nodeutil.StartAsyncRestart(log, restartFn)

	cert, _, err := pkiutil.LoadCertificate(certificates.KubeletCert, "")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to load kubelet certificate: %w", err))
	}

	expirationTimeUNIX := cert.NotAfter.Unix()

	return utils.SyncManualResponseWithSignal(r, readyCh, apiv1.RefreshCertificatesRunResponse{
		ExpirationSeconds: int(expirationTimeUNIX),
	})
}

// getAllCertsForRole returns a slice of all certificate names for a given role.
func getAllCertsForRole(role apiv1.ClusterRole) []string {
	if roleCerts, ok := certsByRole[role]; ok {
		certs := make([]string, 0, len(roleCerts))
		for certName := range roleCerts {
			certs = append(certs, certName)
		}
		slices.Sort(certs)
		return certs
	}
	return nil
}

// Map of worker certificate names to their corresponding CSR definitions.
func getWorkerCSRDefinitions(hostname string) map[string]*csrDefinition {
	return map[string]*csrDefinition{
		CertKubelet: {
			CertName:     CertKubelet,
			CSRBaseName:  "worker-kubelet-serving",
			CommonName:   fmt.Sprintf("system:node:%s", hostname),
			Organization: []string{"system:nodes"},
			Usages:       []certv1.KeyUsage{certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment, certv1.UsageServerAuth},
			SignerName:   "k8sd.io/kubelet-serving",
			NeedsSANs:    true,
		},
		CertKubeletClient: {
			CertName:     CertKubeletClient,
			CSRBaseName:  "worker-kubelet-client",
			CommonName:   fmt.Sprintf("system:node:%s", hostname),
			Organization: []string{"system:nodes"},
			Usages:       []certv1.KeyUsage{certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment, certv1.UsageClientAuth},
			SignerName:   "k8sd.io/kubelet-client",
			NeedsSANs:    false,
		},
		CertProxyClient: {
			CertName:     CertProxyClient,
			CSRBaseName:  "worker-kube-proxy-client",
			CommonName:   "system:kube-proxy",
			Organization: []string{},
			Usages:       []certv1.KeyUsage{certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment, certv1.UsageClientAuth},
			SignerName:   "k8sd.io/kube-proxy-client",
			NeedsSANs:    false,
		},
	}
}

// isCertificateSigningRequestApprovedAndIssued checks if the certificate
// signing request is approved and issued. It returns true if the CSR is
// approved and issued, false if it is pending, and an error if it is denied
// or failed.
func isCertificateSigningRequestApprovedAndIssued(csr *certv1.CertificateSigningRequest) (bool, error) {
	for _, condition := range csr.Status.Conditions {
		if condition.Type == certv1.CertificateApproved && condition.Status == corev1.ConditionTrue {
			return len(csr.Status.Certificate) > 0, nil
		}
		if condition.Type == certv1.CertificateDenied && condition.Status == corev1.ConditionTrue {
			return false, fmt.Errorf("CSR %s was denied: %s", csr.Name, condition.Reason)
		}
		if condition.Type == certv1.CertificateFailed && condition.Status == corev1.ConditionTrue {
			return false, fmt.Errorf("CSR %s failed: %s", csr.Name, condition.Reason)
		}
	}
	return false, nil
}

// verifyCSRAndSetPKI verifies the certificate signing request and sets the
// certificate and key if the CSR is approved.
func verifyCSRAndSetPKI(csr *certv1.CertificateSigningRequest, keyPEM string, certificate, key *string) (bool, error) {
	approved, err := isCertificateSigningRequestApprovedAndIssued(csr)
	if err != nil {
		return false, fmt.Errorf("failed to validate csr: %w", err)
	}

	if !approved {
		return false, nil
	}

	if _, _, err = pkiutil.LoadCertificate(string(csr.Status.Certificate), ""); err != nil {
		return false, fmt.Errorf("failed to load certificate: %w", err)
	}

	*certificate = string(csr.Status.Certificate)
	*key = keyPEM
	return true, nil
}
