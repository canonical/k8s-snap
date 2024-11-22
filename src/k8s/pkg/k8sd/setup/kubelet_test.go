package setup_test

import (
	"net"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

// These values are hard-coded and need to be updated if the
// implementation changes.
var (
	expectedControlPlaneLabels = "node-role.kubernetes.io/control-plane=,node-role.kubernetes.io/worker=,k8sd.io/role=control-plane"
	expectedWorkerLabels       = "node-role.kubernetes.io/worker=,k8sd.io/role=worker"
)

var kubeletTLSCipherSuites = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384"

func mustSetupSnapAndDirectories(t *testing.T, createMock func(*mock.Snap, string)) (s *mock.Snap) {
	g := NewWithT(t)
	dir := t.TempDir()
	s = &mock.Snap{}
	createMock(s, dir)
	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())
	return s
}

func setKubeletMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		KubernetesPKIDir:     filepath.Join(dir, "pki"),
		KubernetesConfigDir:  filepath.Join(dir, "k8s-config"),
		KubeletRootDir:       filepath.Join(dir, "kubelet-root"),
		ServiceArgumentsDir:  filepath.Join(dir, "args"),
		ContainerdSocketDir:  filepath.Join(dir, "containerd-run"),
		ContainerdSocketPath: filepath.Join(dir, "containerd-run", "containerd.sock"),
	}
}

func TestKubelet(t *testing.T) {
	t.Run("ControlPlaneArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil, nil)).To(Succeed())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedControlPlaneLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
			{key: "--cluster-dns", expectedVal: "10.152.1.1"},
			{key: "--cloud-provider", expectedVal: "provider"},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("ControlPlaneWithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		extraArgs := map[string]*string{
			"--cluster-domain": utils.Pointer("override.local"),
			"--cloud-provider": nil, // This should trigger a delete
			"--my-extra-arg":   utils.Pointer("my-extra-val"),
		}
		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil, extraArgs)).To(Succeed())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedControlPlaneLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
			{key: "--cluster-dns", expectedVal: "10.152.1.1"},
			// Overwritten by extraArgs
			{key: "--cluster-domain", expectedVal: "override.local"},
			{key: "--node-ip", expectedVal: "192.168.0.1"},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure that the cloud-provider argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kubelet", "--cloud-provider")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("ControlPlaneArgsNoOptional", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", nil, "", "", "", nil, nil)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedControlPlaneLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("WorkerArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		// Call the kubelet worker setup function
		g.Expect(setup.KubeletWorker(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil)).To(Succeed())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedWorkerLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("WorkerWithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		extraArgs := map[string]*string{
			"--cluster-domain": utils.Pointer("override.local"),
			"--cloud-provider": nil,
		}

		// Call the kubelet worker setup function
		g.Expect(setup.KubeletWorker(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", extraArgs)).To(Succeed())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedWorkerLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
			{key: "--cluster-dns", expectedVal: "10.152.1.1"},
			{key: "--cluster-domain", expectedVal: "override.local"},
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

		// Ensure that the cloud-provider argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kubelet", "--cloud-provider")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())

		// Ensure the kubelet arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("WorkerArgsNoOptional", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		// Call the kubelet worker setup function
		g.Expect(setup.KubeletWorker(s, "dev", nil, "", "", "", nil)).To(Succeed())

		// Ensure the kubelet arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authorization-mode", expectedVal: "Webhook"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook", expectedVal: "true"},
			{key: "--client-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--container-runtime-endpoint", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--containerd", expectedVal: s.Mock.ContainerdSocketPath},
			{key: "--cgroup-driver", expectedVal: "systemd"},
			{key: "--eviction-hard", expectedVal: "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{key: "--fail-swap-on", expectedVal: "false"},
			{key: "--hostname-override", expectedVal: "dev"},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--node-labels", expectedVal: expectedWorkerLabels},
			{key: "--read-only-port", expectedVal: "0"},
			{key: "--register-with-taints", expectedVal: ""},
			{key: "--root-dir", expectedVal: s.Mock.KubeletRootDir},
			{key: "--serialize-image-pulls", expectedVal: "false"},
			{key: "--tls-cipher-suites", expectedVal: kubeletTLSCipherSuites},
			{key: "--tls-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.crt")},
			{key: "--tls-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "kubelet.key")},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kubelet"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("ControlPlaneNoArgsDir", func(t *testing.T) {
		g := NewWithT(t)
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil, nil)).ToNot(Succeed())
	})

	t.Run("WorkerNoArgsDir", func(t *testing.T) {
		g := NewWithT(t)
		s := mustSetupSnapAndDirectories(t, setKubeletMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.KubeletWorker(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil)).ToNot(Succeed())
	})

	t.Run("HostnameOverride", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)
		s.Mock.Hostname = "dev"

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider", nil, nil)).To(Succeed())

		val, err := snaputil.GetServiceArgument(s, "kubelet", "--hostname-override")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(val).To(BeEmpty())
	})

	t.Run("IPv6", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)
		s.Mock.Hostname = "dev"

		// Call the kubelet control plane setup function
		g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("2001:db8::"), "2001:db8::1", "test-cluster.local", "provider", nil, nil)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--cluster-dns", expectedVal: "2001:db8::1"},
			{key: "--node-ip", expectedVal: "2001:db8::"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kubelet", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
	})
}
