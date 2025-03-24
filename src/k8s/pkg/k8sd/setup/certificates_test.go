package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

// TestEnsureK8sDqlitePKI tests the EnsureK8sDqlitePKI function.
func TestEnsureK8sDqlitePKI(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			K8sDqliteStateDir: tempDir,
			UID:               os.Getuid(),
			GID:               os.Getgid(),
		},
	}
	certificates := &pki.K8sDqlitePKI{
		K8sDqliteCert: "dqlite_cert",
		K8sDqliteKey:  "dqlite_key",
	}

	_, err := setup.EnsureK8sDqlitePKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	expectedFiles := []string{
		filepath.Join(tempDir, "cluster.crt"),
		filepath.Join(tempDir, "cluster.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(Not(HaveOccurred()))
	}
}

// TestEnsureControlPlanePKI tests the EnsureControlPlanePKI function.
func TestEnsureControlPlanePKI(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir: tempDir,
			UID:              os.Getuid(),
			GID:              os.Getgid(),
		},
	}
	certificates := &pki.ControlPlanePKI{
		CACert:                     "ca_cert",
		CAKey:                      "ca_key",
		ClientCACert:               "client_ca_cert",
		ClientCAKey:                "client_ca_key",
		FrontProxyCACert:           "front_proxy_ca_cert",
		FrontProxyCAKey:            "front_proxy_ca_key",
		FrontProxyClientCert:       "front_proxy_client_cert",
		FrontProxyClientKey:        "front_proxy_client_key",
		APIServerCert:              "apiserver_cert",
		APIServerKey:               "apiserver_key",
		APIServerKubeletClientCert: "apiserver_kubelet_client_cert",
		APIServerKubeletClientKey:  "apiserver_kubelet_client_key",
		KubeletCert:                "kubelet_cert",
		KubeletKey:                 "kubelet_key",
		ServiceAccountKey:          "serviceaccount_key",
	}

	_, err := setup.EnsureControlPlanePKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	expectedFiles := []string{
		filepath.Join(tempDir, "apiserver-kubelet-client.crt"),
		filepath.Join(tempDir, "apiserver-kubelet-client.key"),
		filepath.Join(tempDir, "apiserver.crt"),
		filepath.Join(tempDir, "apiserver.key"),
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client-ca.crt"),
		filepath.Join(tempDir, "client-ca.key"),
		filepath.Join(tempDir, "front-proxy-ca.crt"),
		filepath.Join(tempDir, "front-proxy-client.crt"),
		filepath.Join(tempDir, "front-proxy-client.key"),
		filepath.Join(tempDir, "kubelet.crt"),
		filepath.Join(tempDir, "kubelet.key"),
		filepath.Join(tempDir, "serviceaccount.key"),
		filepath.Join(tempDir, "ca.key"),
		filepath.Join(tempDir, "front-proxy-ca.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(Not(HaveOccurred()))
	}
}

// TestEnsureWorkerPKI tests the EnsureWorkerPKI function.
func TestEnsureWorkerPKI(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir: tempDir,
			UID:              os.Getuid(),
			GID:              os.Getgid(),
		},
	}
	certificates := &pki.WorkerNodePKI{
		CACert:       "ca_cert",
		ClientCACert: "client_ca_cert",
		KubeletCert:  "kubelet_cert",
		KubeletKey:   "kubelet_key",
	}

	_, err := setup.EnsureWorkerPKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	expectedFiles := []string{
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client-ca.crt"),
		filepath.Join(tempDir, "kubelet.crt"),
		filepath.Join(tempDir, "kubelet.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(Not(HaveOccurred()))
	}
}

func TestExtDatastorePKI(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			EtcdPKIDir: tempDir,
			UID:        os.Getuid(),
			GID:        os.Getgid(),
		},
	}
	certificates := &pki.ExternalDatastorePKI{
		DatastoreCACert:     "ca_cert",
		DatastoreClientKey:  "client_key",
		DatastoreClientCert: "client_cert",
	}

	_, err := setup.EnsureExtDatastorePKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	expectedFiles := []string{
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client.key"),
		filepath.Join(tempDir, "client.crt"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(Not(HaveOccurred()))
	}
}

// Check that a file passed to Ensure*PKI is deleted if the corresponding
// certificate content is empty.
func TestEmptyCert(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			K8sDqliteStateDir: tempDir,
			UID:               os.Getuid(),
			GID:               os.Getgid(),
		},
	}

	expectedFiles := []string{
		filepath.Join(tempDir, "cluster.crt"),
		filepath.Join(tempDir, "cluster.key"),
	}

	certificates := &pki.K8sDqlitePKI{
		K8sDqliteCert: "dqlite-cert",
		K8sDqliteKey:  "dqlite-key",
	}

	// Should create files
	_, err := setup.EnsureK8sDqlitePKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	certificates = &pki.K8sDqlitePKI{
		K8sDqliteCert: "",
		K8sDqliteKey:  "",
	}

	// Should delete files
	_, err = setup.EnsureK8sDqlitePKI(mock, certificates)
	g.Expect(err).To(Not(HaveOccurred()))

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(HaveOccurred())
	}
}

func TestReadControlPlanePKI(t *testing.T) {
	g := NewWithT(t)
	cases := []struct {
		name             string
		readManaged      bool
		removeCAKeyFiles bool
	}{
		{"Managed", true, false},
		{"External", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pkiDir := t.TempDir()
			configDir := t.TempDir()

			mock := &mock.Snap{
				Mock: mock.Mock{
					KubernetesPKIDir:    pkiDir,
					KubernetesConfigDir: configDir,
					UID:                 os.Getuid(),
					GID:                 os.Getgid(),
				},
			}

			orig := &pki.ControlPlanePKI{
				CACert:                          "ca_cert_val",
				CAKey:                           "ca_key_val",
				ClientCACert:                    "client_ca_cert_val",
				ClientCAKey:                     "client_ca_key_val",
				FrontProxyCACert:                "front_proxy_ca_cert_val",
				FrontProxyCAKey:                 "front_proxy_ca_key_val",
				APIServerCert:                   "apiserver_cert_val",
				APIServerKey:                    "apiserver_key_val",
				KubeletCert:                     "kubelet_cert_val",
				KubeletKey:                      "kubelet_key_val",
				FrontProxyClientCert:            "front_proxy_client_cert_val",
				FrontProxyClientKey:             "front_proxy_client_key_val",
				APIServerKubeletClientCert:      "apiserver_kubelet_client_cert_val",
				APIServerKubeletClientKey:       "apiserver_kubelet_client_key_val",
				ServiceAccountKey:               "serviceaccount_key_val",
				AdminClientCert:                 "admin_cert_val",
				AdminClientKey:                  "admin_key_val",
				KubeControllerManagerClientCert: "controller_cert_val",
				KubeControllerManagerClientKey:  "controller_key_val",
				KubeSchedulerClientCert:         "scheduler_cert_val",
				KubeSchedulerClientKey:          "scheduler_key_val",
				KubeProxyClientCert:             "proxy_cert_val",
				KubeProxyClientKey:              "proxy_key_val",
				KubeletClientCert:               "kubelet_cert_val",
				KubeletClientKey:                "kubelet_key_val",
			}

			_, err := setup.EnsureControlPlanePKI(mock, orig)
			g.Expect(err).ToNot(HaveOccurred())

			err = setup.SetupControlPlaneKubeconfigs(mock.KubernetesConfigDir(), "127.0.0.1", 8443, *orig)
			g.Expect(err).ToNot(HaveOccurred())

			if tc.removeCAKeyFiles {
				os.Remove(filepath.Join(pkiDir, "ca.key"))
				os.Remove(filepath.Join(pkiDir, "client-ca.key"))
				os.Remove(filepath.Join(pkiDir, "front-proxy-ca.key"))
			}

			readCerts := &pki.ControlPlanePKI{}

			err = setup.ReadControlPlanePKI(mock, readCerts, tc.readManaged)
			g.Expect(err).ToNot(HaveOccurred())

			// NOTE: Always read CA certificate files and Service Account.
			g.Expect(readCerts.CACert).To(Equal("ca_cert_val"))
			g.Expect(readCerts.ClientCACert).To(Equal("client_ca_cert_val"))
			g.Expect(readCerts.FrontProxyCACert).To(Equal("front_proxy_ca_cert_val"))
			g.Expect(readCerts.ServiceAccountKey).To(Equal("serviceaccount_key_val"))

			if !tc.readManaged {
				g.Expect(readCerts.CAKey).To(BeEmpty())
				g.Expect(readCerts.ClientCAKey).To(BeEmpty())
				g.Expect(readCerts.FrontProxyCAKey).To(BeEmpty())
			}

			g.Expect(readCerts.APIServerCert).To(Equal("apiserver_cert_val"))
			g.Expect(readCerts.APIServerKey).To(Equal("apiserver_key_val"))
			g.Expect(readCerts.KubeletCert).To(Equal("kubelet_cert_val"))
			g.Expect(readCerts.KubeletKey).To(Equal("kubelet_key_val"))
			g.Expect(readCerts.FrontProxyClientCert).To(Equal("front_proxy_client_cert_val"))
			g.Expect(readCerts.FrontProxyClientKey).To(Equal("front_proxy_client_key_val"))
			g.Expect(readCerts.APIServerKubeletClientCert).To(Equal("apiserver_kubelet_client_cert_val"))
			g.Expect(readCerts.APIServerKubeletClientKey).To(Equal("apiserver_kubelet_client_key_val"))

			g.Expect(readCerts.AdminClientCert).To(Equal("admin_cert_val"))
			g.Expect(readCerts.AdminClientKey).To(Equal("admin_key_val"))
			g.Expect(readCerts.KubeControllerManagerClientCert).To(Equal("controller_cert_val"))
			g.Expect(readCerts.KubeControllerManagerClientKey).To(Equal("controller_key_val"))
			g.Expect(readCerts.KubeSchedulerClientCert).To(Equal("scheduler_cert_val"))
			g.Expect(readCerts.KubeSchedulerClientKey).To(Equal("scheduler_key_val"))
			g.Expect(readCerts.KubeProxyClientCert).To(Equal("proxy_cert_val"))
			g.Expect(readCerts.KubeProxyClientKey).To(Equal("proxy_key_val"))
			g.Expect(readCerts.KubeletClientCert).To(Equal("kubelet_cert_val"))
			g.Expect(readCerts.KubeletClientKey).To(Equal("kubelet_key_val"))
		})
	}
}
