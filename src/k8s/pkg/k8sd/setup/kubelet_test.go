package setup_test

import (
	"net"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

// These values are hard-coded and need to be updated if the
// implementation changes.
var expectedControlPlaneLabels = "node-role.kubernetes.io/control-plane=,node-role.kubernetes.io/worker="
var expectedWorkerLabels = "node-role.kubernetes.io/worker="

var kubeletTLSCipherSuites = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384"

func mustSetupSnapAndDirectories(t *testing.T, createMock func(*mock.Snap, string)) (s *mock.Snap, dir string) {
	g := NewWithT(t)
	dir = t.TempDir()
	s = &mock.Snap{}
	createMock(s, dir)
	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())
	return s, dir
}

func mustReturnMockForKubelet(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		KubernetesPKIDir:    path.Join(dir, "pki"),
		KubernetesConfigDir: path.Join(dir, "k8s-config"),
		KubeletRootDir:      path.Join(dir, "kubelet-root"),
		ServiceArgumentsDir: path.Join(dir, "args"),
		ContainerdSocketDir: path.Join(dir, "containerd-run"),
	}
}

func TestKubelet(t *testing.T) {
	t.Run("ControlPlaneArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubelet)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).To(BeNil())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--cert-dir", expectedVal: s.Mock.KubernetesPKIDir},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--containerd", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedControlPlaneLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--cloud-provider", expectedVal: "provider"},
			{key: "--cluster-dns", expectedVal: "10.152.1.1"},
			{key: "--cluster-domain", expectedVal: "test-cluster.local"},
			{key: "--node-ip", expectedVal: "192.168.0.1"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("ControlPlaneArgsNoOptional", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubelet)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", nil, "", "", "")).To(BeNil())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--cert-dir", expectedVal: s.Mock.KubernetesPKIDir},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--containerd", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedControlPlaneLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("WorkerArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubelet)

		// Call the kubelet worker setup function
		g.Expect(setup.KubeletWorker(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).To(BeNil())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--cert-dir", expectedVal: s.Mock.KubernetesPKIDir},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--containerd", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedWorkerLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--cloud-provider", expectedVal: "provider"},
			{key: "--cluster-dns", expectedVal: "10.152.1.1"},
			{key: "--cluster-domain", expectedVal: "test-cluster.local"},
			{key: "--node-ip", expectedVal: "192.168.0.1"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("WorkerArgsNoOptional", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubelet)

		// Call the kubelet worker setup function
		g.Expect(setup.KubeletWorker(s, "dev", nil, "", "", "")).To(BeNil())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--cert-dir", expectedVal: s.Mock.KubernetesPKIDir},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--containerd", expectedVal: path.Join(s.Mock.ContainerdSocketDir, "containerd.sock")},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedWorkerLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("MissingServiceArgumentsDir", func(t *testing.T) {
		g := NewWithT(t)
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubelet)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.KubeletControlPlane(s, "", nil, "", "", "")).ToNot(Succeed())
		g.Expect(setup.KubeletWorker(s, "", nil, "", "", "")).ToNot(Succeed())

		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).ToNot(Succeed())
		g.Expect(setup.KubeletWorker(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).ToNot(Succeed())
	})
}
