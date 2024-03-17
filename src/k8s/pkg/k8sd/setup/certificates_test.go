package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
)

// TestEnsureK8sDqlitePKI tests the EnsureK8sDqlitePKI function.
func TestEnsureK8sDqlitePKI(t *testing.T) {
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

	err := setup.EnsureK8sDqlitePKI(mock, certificates)
	if err != nil {
		t.Fatalf("EnsureK8sDqlitePKI returned unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(tempDir, "cluster.crt"),
		filepath.Join(tempDir, "cluster.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		if err != nil {
			t.Errorf("Expected file %q is missing: %v", file, err)
		}
	}
}

// TestEnsureControlPlanePKI tests the EnsureControlPlanePKI function.
func TestEnsureControlPlanePKI(t *testing.T) {
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

	err := setup.EnsureControlPlanePKI(mock, certificates)
	if err != nil {
		t.Fatalf("EnsureControlPlanePKI returned unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(tempDir, "apiserver-kubelet-client.crt"),
		filepath.Join(tempDir, "apiserver-kubelet-client.key"),
		filepath.Join(tempDir, "apiserver.crt"),
		filepath.Join(tempDir, "apiserver.key"),
		filepath.Join(tempDir, "ca.crt"),
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
		if err != nil {
			t.Errorf("Expected file %q is missing: %v", file, err)
		}
	}
}

// TestEnsureWorkerPKI tests the EnsureWorkerPKI function.
func TestEnsureWorkerPKI(t *testing.T) {
	tempDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir: tempDir,
			UID:              os.Getuid(),
			GID:              os.Getgid(),
		},
	}
	certificates := &pki.WorkerNodePKI{
		CACert:      "ca_cert",
		KubeletCert: "kubelet_cert",
		KubeletKey:  "kubelet_key",
	}

	err := setup.EnsureWorkerPKI(mock, certificates)
	if err != nil {
		t.Fatalf("EnsureWorkerPKI returned unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(tempDir, "ca.crt"),
		filepath.Join(tempDir, "kubelet.crt"),
		filepath.Join(tempDir, "kubelet.key"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		if err != nil {
			t.Errorf("Expected file %q is missing: %v", file, err)
		}
	}
}
