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
	g.Expect(err).To(BeNil())

	expectedFiles := []string{
		filepath.Join(tempDir, "cluster.crt"),
		filepath.Join(tempDir, "cluster.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(BeNil())
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
	g.Expect(err).To(BeNil())

	expectedFiles := []string{
		filepath.Join(tempDir, "apiserver-kubelet-client.crt"),
		filepath.Join(tempDir, "apiserver-kubelet-client.key"),
		filepath.Join(tempDir, "apiserver.crt"),
		filepath.Join(tempDir, "apiserver.key"),
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client-ca.crt"),
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
		g.Expect(err).To(BeNil())
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
	g.Expect(err).To(BeNil())

	expectedFiles := []string{
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client-ca.crt"),
		filepath.Join(tempDir, "kubelet.crt"),
		filepath.Join(tempDir, "kubelet.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(BeNil())
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
	g.Expect(err).To(BeNil())

	expectedFiles := []string{
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "client.key"),
		filepath.Join(tempDir, "client.crt"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(BeNil())
	}
}

func TestEtcdPKI(t *testing.T) {
	g := NewWithT(t)
	etcdPKI := t.TempDir()
	kubePKI := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir: kubePKI,
			EtcdPKIDir:       etcdPKI,
			UID:              os.Getuid(),
			GID:              os.Getgid(),
		},
	}
	certificates := &pki.EtcdPKI{
		CACert:                  "ca_cert",
		CAKey:                   "ca_key",
		ServerCert:              "server_cert",
		ServerKey:               "server_key",
		ServerPeerCert:          "server_peer_cert",
		ServerPeerKey:           "server_peer_key",
		KubeAPIServerClientCert: "client_cert",
		KubeAPIServerClientKey:  "client_key",
	}

	_, err := setup.EnsureEtcdPKI(mock, certificates)
	g.Expect(err).To(BeNil())

	expectedFiles := []string{
		filepath.Join(etcdPKI, "ca.crt"),
		filepath.Join(etcdPKI, "server.crt"),
		filepath.Join(etcdPKI, "server.key"),
		filepath.Join(etcdPKI, "peer.crt"),
		filepath.Join(etcdPKI, "peer.key"),
		filepath.Join(kubePKI, "apiserver-etcd-client.crt"),
		filepath.Join(kubePKI, "apiserver-etcd-client.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).To(BeNil())
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
	g.Expect(err).To(BeNil())

	certificates = &pki.K8sDqlitePKI{
		K8sDqliteCert: "",
		K8sDqliteKey:  "",
	}

	// Should delete files
	_, err = setup.EnsureK8sDqlitePKI(mock, certificates)
	g.Expect(err).To(BeNil())

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		g.Expect(err).NotTo(BeNil())
	}
}
