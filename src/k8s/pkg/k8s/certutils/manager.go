package certutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/k8s/utils"
)

// CertificateManager contains and manages certificates that is used by the k8s node.
type CertificateManager struct {
	hostname              string
	defaultIp             net.IP
	CA                    *CertKeyPair
	FrontProxyCa          *CertKeyPair
	FrontProxyClient      *CertKeyPair
	K8sDqlite             *CertKeyPair
	KubeAdmin             *CertKeyPair
	KubeApiserver         *CertKeyPair
	KubeControllerManager *CertKeyPair
	KubeProxy             *CertKeyPair
	KubeScheduler         *CertKeyPair
	Kubelet               *CertKeyPair
	KubeletClient         *CertKeyPair
}

// NewCertificateManager returns a new CertificateManager.
func NewCertificateManager() (*CertificateManager, error) {
	cm := &CertificateManager{}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	cm.hostname = hostname

	defaultIp, err := utils.GetDefaultIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get default ip: %w", err)
	}
	cm.defaultIp = defaultIp

	return cm, nil
}

// GenerateServerCerts generates all the necessary certificates to be used.
func (cm *CertificateManager) GenerateServerCerts() (err error) {
	err = cm.generateFrontProxyClient()
	if err != nil {
		return err
	}
	cm.generateK8sDqlite()
	if err != nil {
		return err
	}
	cm.generateKubeApiserver()
	if err != nil {
		return err
	}
	cm.generateKubelet()
	if err != nil {
		return err
	}
	cm.generateKubeletClient()
	if err != nil {
		return err
	}
	return nil
}

// GenerateClientCerts generates certificates that can be used to communicate with the apiserver through certificate authentication.
func (cm *CertificateManager) GenerateClientCerts() (err error) {
	cm.generateKubeAdmin()
	if err != nil {
		return err
	}
	cm.generateKubeControllerManager()
	if err != nil {
		return err
	}
	cm.generateKubeProxy()
	if err != nil {
		return err
	}
	cm.generateKubeScheduler()
	if err != nil {
		return err
	}
	return nil
}

// GenerateServiceAccountKey generates a private key for the service account.
func (cm *CertificateManager) GenerateServiceAccountKey() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate service account key: %w", err)
	}

	ckp, err := NewCertKeyPair(nil, key)
	if err != nil {
		return fmt.Errorf("failed to create cert-key pair: %w", err)
	}

	err = ckp.SavePrivateKey(filepath.Join(KubePkiPath, "serviceaccount.key"))
	if err != nil {
		return fmt.Errorf("failed to save service account key: %w", err)
	}

	return nil

}

// GenerateCA generates the kubernetes wide certificate authority.
func (cm *CertificateManager) GenerateCA() error {
	templ, err := NewCATemplate(
		pkix.Name{
			CommonName: "kubernetes-ca",
		},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"ca",
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate authority template: %w", err)
	}

	cert, err := templ.SignAndSave(true, nil)
	if err != nil {
		return fmt.Errorf("failed to sign and save certificate authority: %w", err)
	}

	cm.CA = cert
	return nil
}

// GenerateFrontProxyCA generates the certificate authority for the front-proxy.
func (cm *CertificateManager) GenerateFrontProxyCA() error {
	templ, err := NewCATemplate(
		pkix.Name{
			CommonName: "kubernetes-front-proxy-ca",
		},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"front-proxy-ca",
	)
	if err != nil {
		return fmt.Errorf("failed to create front proxy certificate authority template: %w", err)
	}

	cert, err := templ.SignAndSave(true, nil)
	if err != nil {
		return fmt.Errorf("failed to sign and save front proxy certificate authority: %w", err)
	}

	cm.FrontProxyCa = cert
	return nil
}

func (cm *CertificateManager) generateFrontProxyClient() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			Organization:       []string{"Canonical"},
			OrganizationalUnit: []string{"Canonical"},
			Country:            []string{"GB"},
			Province:           []string{""},
			Locality:           []string{"Canonical"},
			StreetAddress:      []string{"Canonical"},
			CommonName:         "front-proxy-client",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"front-proxy-client",
	)
	if err != nil {
		return fmt.Errorf("failed to create front proxy client certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.FrontProxyCa)
	if err != nil {
		return fmt.Errorf("failed to sign and save front proxy client certificate: %w", err)
	}

	cm.FrontProxyClient = cert
	return nil
}

func (cm *CertificateManager) generateK8sDqlite() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			Organization:       []string{"Canonical"},
			OrganizationalUnit: []string{"Canonical"},
			Country:            []string{"GB"},
			Province:           []string{""},
			Locality:           []string{"Canonical"},
			StreetAddress:      []string{"Canonical"},
			CommonName:         "k8s",
		},
		[]string{cm.hostname},
		[]net.IP{net.IPv4(127, 0, 0, 1)},
		time.Now().AddDate(10, 0, 0),
		2048,
		K8sDqlitePkiPath,
		"cluster",
	)
	if err != nil {
		return fmt.Errorf("failed to create k8s-dqlite certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(true, nil)
	if err != nil {
		return fmt.Errorf("failed to sign and save k8s-dqlite certificate: %w", err)
	}

	cm.K8sDqlite = cert
	return nil
}

func (cm *CertificateManager) generateKubeApiserver() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			CommonName: "kube-apiserver",
		},
		[]string{"localhost", "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", cm.hostname},
		[]net.IP{net.IPv4(127, 0, 0, 1), net.IPv4(10, 152, 183, 1), cm.defaultIp},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"apiserver",
	)
	if err != nil {
		return fmt.Errorf("failed to create kube-apiserver certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kube-apiserver certificate: %w", err)
	}

	cm.KubeApiserver = cert
	return nil
}

func (cm *CertificateManager) generateKubeAdmin() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			Organization: []string{"system:masters"},
			CommonName:   "kubernetes-admin",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"admin",
	)
	if err != nil {
		return fmt.Errorf("failed to create kube-admin certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kube-admin certificate: %w", err)
	}

	cm.KubeAdmin = cert
	return nil
}

func (cm *CertificateManager) generateKubeControllerManager() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			CommonName: "system:kube-controller-manager",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"controller-manager",
	)
	if err != nil {
		return fmt.Errorf("failed to create kube-controller-manager certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kube-controller-manager certificate: %w", err)
	}

	cm.KubeControllerManager = cert
	return nil
}

func (cm *CertificateManager) generateKubeProxy() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			CommonName: "system:kube-proxy",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"proxy",
	)
	if err != nil {
		return fmt.Errorf("failed to create kube-proxy certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kube-proxy certificate: %w", err)
	}

	cm.KubeProxy = cert
	return nil
}

func (cm *CertificateManager) generateKubeScheduler() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			CommonName: "system:kube-scheduler",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"scheduler",
	)
	if err != nil {
		return fmt.Errorf("failed to create kube-scheduler certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kube-scheduler certificate: %w", err)
	}

	cm.KubeScheduler = cert
	return nil
}

func (cm *CertificateManager) generateKubelet() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			Organization: []string{"system:nodes"},
			CommonName:   fmt.Sprintf("system:node:%s", cm.hostname),
		},
		[]string{cm.hostname},
		[]net.IP{net.IPv4(127, 0, 0, 1), cm.defaultIp},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"kubelet",
	)
	if err != nil {
		return fmt.Errorf("failed to create kubelet certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kubelet certificate: %w", err)
	}

	cm.Kubelet = cert
	return nil
}

func (cm *CertificateManager) generateKubeletClient() error {
	templ, err := NewCertificateTemplate(
		pkix.Name{
			Organization: []string{"system:masters"},
			CommonName:   "kube-apiserver-kubelet-client",
		},
		[]string{},
		[]net.IP{},
		time.Now().AddDate(10, 0, 0),
		2048,
		KubePkiPath,
		"apiserver-kubelet-client",
	)
	if err != nil {
		return fmt.Errorf("failed to create kubelet client certificate template: %w", err)
	}

	cert, err := templ.SignAndSave(false, cm.CA)
	if err != nil {
		return fmt.Errorf("failed to sign and save kubelet client certificate: %w", err)
	}

	cm.KubeletClient = cert
	return nil
}
