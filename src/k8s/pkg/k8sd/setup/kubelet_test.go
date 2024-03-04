package setup_test

import (
	"path"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestKubelet(t *testing.T) {
	g := NewWithT(t)

	dir := t.TempDir()

	s := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir:    path.Join(dir, "pki"),
			KubernetesConfigDir: path.Join(dir, "k8s-config"),
			KubeletRootDir:      path.Join(dir, "kubelet-root"),
			ServiceArgumentsDir: path.Join(dir, "args"),
			ContainerdSocketDir: path.Join(dir, "containerd-run"),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())

	g.Expect(setup.KubeletControlPlane(s, "dev", nil, "10.152.1.1", "test-cluster.local", "")).To(BeNil())

	t.Run("Args", func(t *testing.T) {
		kubeletControlPlaneLabels := []string{"node-role.kubernetes.io/control-plane="}
		kubeletWorkerLabels := []string{"node-role.kubernetes.io/worker="}
		expectedLabels := strings.Join(append(kubeletControlPlaneLabels, kubeletWorkerLabels...), ",")

		var kubeletTLSCipherSuites = []string{
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
			"TLS_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_RSA_WITH_AES_256_GCM_SHA384",
		}

		tests := []struct {
			key         string
			expectedVal string
		}{
			{"--anonymous-auth", "false"},
			{"--authentication-token-webhook", "true"},
			{"--cert-dir", s.Mock.KubernetesPKIDir},
			{"--client-ca-file", path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{"--container-runtime-endpoint", path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{"--containerd", path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{"--eviction-hard", "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{"--fail-swap-on", "false"},
			{"--hostname-override", "dev"},
			{"--kubeconfig", path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{"--node-labels", expectedLabels},
			{"--read-only-port", "0"},
			{"--register-with-taints", ""},
			{"--root-dir", s.Mock.KubeletRootDir},
			{"--serialize-image-pulls", "false"},
			{"--tls-cipher-suites", strings.Join(kubeletTLSCipherSuites, ",")},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
	})
}
