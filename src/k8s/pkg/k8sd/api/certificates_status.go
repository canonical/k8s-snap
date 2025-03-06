package api

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"k8s.io/client-go/tools/clientcmd"
)

func (e *Endpoints) getCertificatesStatus(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()
	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}
	if isWorker {
		return getCertsStatusWorker(s, r, snap)
	}
	return getCertsStatusControlPlane(s, r, snap)
}

// getCertsStatusControlPlane collects certificate status information for
// control plane nodes. It reads control plane certificates, kubeconfig
// certificates, and certificate authority statuses.
func getCertsStatusControlPlane(s state.State, r *http.Request, snap snap.Snap) response.Response {
	clusterConfig, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve cluster configuration: %w", err))
	}
	authorities, err := readCertificateAuthorities(&clusterConfig)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read certificates authorities: %w", err))
	}

	certsNames := []string{"apiserver", "apiserver-kubelet-client", "front-proxy-client", "kubelet"}
	nodeCerts, err := loadCertificateStatusesFromDir(snap.KubernetesPKIDir(), certsNames)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read node certificates: %w", err))
	}

	kubeConfigs := []string{"admin.conf", "controller.conf", "kubelet.conf", "proxy.conf", "scheduler.conf"}
	kubeConfigCerts, err := readKubeconfigCertificates(snap.KubernetesConfigDir(), kubeConfigs)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read kubeconfig certificates: %w", err))
	}

	var certificates []apiv1.CertificateStatus
	certificates = append(certificates, nodeCerts...)
	certificates = append(certificates, kubeConfigCerts...)

	if clusterConfig.Datastore.GetType() == "external" {
		dataStoreCerts, err := loadCertificateStatusesFromDir(snap.EtcdPKIDir(), []string{"client"})
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to read datastore certificates: %w", err))
		}
		certificates = append(certificates, dataStoreCerts...)
	}

	updateExternallyManaged(authorities, certificates)
	return response.SyncResponse(true, apiv1.CertificatesStatusResponse{
		Certificates:           certificates,
		CertificateAuthorities: authorities,
	})
}

// getCertsStatusWorker collects certificate status information for worker
// nodes. It reads worker certificates and kubeconfig certificates.
func getCertsStatusWorker(s state.State, r *http.Request, snap snap.Snap) response.Response {
	certsNames := []string{"kubelet"}
	nodeCerts, err := loadCertificateStatusesFromDir(snap.KubernetesPKIDir(), certsNames)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read node certificates: %w", err))
	}

	kubeConfigs := []string{"kubelet.conf", "proxy.conf"}
	kubeConfigCerts, err := readKubeconfigCertificates(snap.KubernetesConfigDir(), kubeConfigs)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to read kubeconfig certificates: %w", err))
	}

	var certificates []apiv1.CertificateStatus
	certificates = append(certificates, nodeCerts...)
	certificates = append(certificates, kubeConfigCerts...)

	// NOTE: Worker certificates are inherently externally managed.
	for i := range certificates {
		cert := &certificates[i]
		cert.ExternallyManaged = true
	}

	return response.SyncResponse(true, apiv1.CertificatesStatusResponse{
		Certificates:           certificates,
		CertificateAuthorities: []apiv1.CertificateAuthorityStatus{},
	})
}

// readKubeconfigCertificates reads the client certificates from kubeconfig
// files located in the specified directory. It returns a slice of
// CertificateStatus for each valid kubeconfig file.
func readKubeconfigCertificates(kubeconfigDir string, configs []string) ([]apiv1.CertificateStatus, error) {
	certificates := make([]apiv1.CertificateStatus, 0, len(configs))

	for _, config := range configs {
		kubeConfig, err := clientcmd.LoadFromFile(filepath.Join(kubeconfigDir, config))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return certificates, fmt.Errorf("failed to load kubeconfig %s: %w", config, err)
		}

		authInfo, exists := kubeConfig.AuthInfos["k8s-user"]
		if !exists {
			return certificates, fmt.Errorf("user 'k8s-user' not found in kubeconfig %s", config)
		}

		if authInfo.ClientCertificateData == nil {
			return certificates, fmt.Errorf("no client certificate data found in kubeconfig %s", config)
		}

		cert, _, err := pkiutil.LoadCertificate(string(authInfo.ClientCertificateData), "")
		if err != nil {
			return certificates, fmt.Errorf("failed to load certificate data from kubeconfig %s: %w", config, err)
		}

		certificates = append(certificates, apiv1.CertificateStatus{
			Name:                 config,
			Expires:              cert.NotAfter.Format(time.RFC3339),
			CertificateAuthority: cert.Issuer.CommonName,
		})
	}
	return certificates, nil
}

// readCertificateAuthorities loads certificate authority information from the
// given cluster configuration.
func readCertificateAuthorities(clusterConfig *types.ClusterConfig) ([]apiv1.CertificateAuthorityStatus, error) {
	cas := make([]apiv1.CertificateAuthorityStatus, 0, 3)

	loadAndAppend := func(certPath, keyPath, name string) error {
		cert, key, err := pkiutil.LoadCertificate(certPath, keyPath)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", name, err)
		}
		cas = append(cas, apiv1.CertificateAuthorityStatus{
			Name:              cert.Subject.CommonName,
			Expires:           cert.NotAfter.Format(time.RFC3339),
			ExternallyManaged: key == nil,
		})
		return nil
	}

	casList := []struct {
		CertPEM string
		KeyPEM  string
		Name    string
	}{
		{clusterConfig.Certificates.GetCACert(), clusterConfig.Certificates.GetCAKey(), "CA"},
		{clusterConfig.Certificates.GetClientCACert(), clusterConfig.Certificates.GetClientCAKey(), "Client CA"},
		{clusterConfig.Certificates.GetFrontProxyCACert(), clusterConfig.Certificates.GetFrontProxyCAKey(), "Front Proxy CA"},
	}

	for _, ca := range casList {
		if err := loadAndAppend(ca.CertPEM, ca.KeyPEM, ca.Name); err != nil {
			return cas, err
		}
	}

	return cas, nil
}

// loadCertificateStatusesFromDir loads the certificate status information for
// the specified certificate names from the given directory. For each
// certificate name, it reads the certificate and key pair using
// loadCertificatePairFromDir, then builds a CertificateStatus that includes
// the certificate's expiration date and issuer information.
func loadCertificateStatusesFromDir(baseDir string, certNames []string) ([]apiv1.CertificateStatus, error) {
	var certs []apiv1.CertificateStatus
	for _, certName := range certNames {
		cert, _, err := loadCertificatePairFromDir(baseDir, certName)
		if err != nil {
			return nil, fmt.Errorf("failed to extract information from certificate %s: %w", certName, err)
		}
		certs = append(certs, apiv1.CertificateStatus{
			Name:                 certName,
			Expires:              cert.NotAfter.Format(time.RFC3339),
			CertificateAuthority: cert.Issuer.CommonName,
		})
	}
	return certs, nil
}

// loadCertificatePairFromDir reads the certificate and corresponding private
// key files for the given certificate name from the specified directory. It
// expects the files to be named "<name>.crt" and "<name>.key".
func loadCertificatePairFromDir(baseDir string, name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certBytes, err := os.ReadFile(filepath.Join(baseDir, fmt.Sprintf("%s.crt", name)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s.crt: %w", name, err)
	}
	keyBytes, err := os.ReadFile(filepath.Join(baseDir, fmt.Sprintf("%s.key", name)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s.key: %w", name, err)
	}
	cert, key, err := pkiutil.LoadCertificate(string(certBytes), string(keyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s certificate: %w", name, err)
	}
	return cert, key, nil
}

// updateExternallyManaged updates the ExternallyManaged field for each
// certificate in the provided slice. It matches each certificate's
// CertificateAuthority against the list of certificate authorities and,
// if a match is found, uses the authority's ExternallyManaged value.
func updateExternallyManaged(authorities []apiv1.CertificateAuthorityStatus, certificates []apiv1.CertificateStatus) {
	for i := range certificates {
		cert := &certificates[i]
		for _, authority := range authorities {
			if authority.Name == cert.CertificateAuthority {
				cert.ExternallyManaged = authority.ExternallyManaged
				break
			}
		}
	}
}
