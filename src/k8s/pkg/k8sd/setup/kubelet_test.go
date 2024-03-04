package setup_test

import (
	"net"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

// These values are hard-coded and need to be updated if the
// implementation changes.
var kubeletControlPlaneLabels = []string{"node-role.kubernetes.io/control-plane="}
var kubeletWorkerLabels = []string{"node-role.kubernetes.io/worker="}
var expectedLabels = strings.Join(append(kubeletControlPlaneLabels, kubeletWorkerLabels...), ",")

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

func testFixture(t *testing.T) (s *mock.Snap, dir string) {
	g := NewWithT(t)

	dir = t.TempDir()

	s = &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir:    path.Join(dir, "pki"),
			KubernetesConfigDir: path.Join(dir, "k8s-config"),
			KubeletRootDir:      path.Join(dir, "kubelet-root"),
			ServiceArgumentsDir: path.Join(dir, "args"),
			ContainerdSocketDir: path.Join(dir, "containerd-run"),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())

	return
}

func TestKubelet(t *testing.T) {
	t.Run("Setup control plane with all possible arguments", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, dir := testFixture(t)
		defer os.RemoveAll(dir)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.IP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).To(BeNil())

		// Ensure the kubelet arguments file has the expected arguments and values
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
			{"--cloud-provider", "provider"},
			{"--cluster-dns", "10.152.1.1"},
			{"--cluster-domain", "test-cluster.local"},
			{"--node-ip", net.IP("192.168.0.1").String()},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).To(BeNil())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("Setup control plane with all optional arguments missing", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, dir := testFixture(t)
		defer os.RemoveAll(dir)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", nil, "", "", "")).To(BeNil())

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

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).To(BeNil())
		g.Expect(len(args)).To(Equal(len(tests)))
	})
}
