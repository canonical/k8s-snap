package setup_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

// testFixture prepares and returns a test environment setup.
func testFixture(t *testing.T) *mock.Snap {
	g := NewWithT(t)
	dir := t.TempDir()

	// Ensure the mock CNI binary can be written.
	require.NoError(t, os.WriteFile(path.Join(dir, "mockcni"), []byte("echo hi"), 0600))

	s := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir:            path.Join(dir, "pki"),
			KubernetesConfigDir:         path.Join(dir, "k8s-config"),
			KubeletRootDir:              path.Join(dir, "kubelet-root"),
			ContainerdSocketDir:         path.Join(dir, "containerd-run"),
			ServiceArgumentsDir:         path.Join(dir, "args"),
			ContainerdConfigDir:         path.Join(dir, "containerd"),
			ContainerdRootDir:           path.Join(dir, "containerd-root"),
			ContainerdRegistryConfigDir: path.Join(dir, "containerd-registries"),
			ContainerdStateDir:          path.Join(dir, "containerd-state"),
			ContainerdExtraConfigDir:    path.Join(dir, "containerd-confd"),
			CNIBinDir:                   path.Join(dir, "opt-cni-bin"),
			CNIConfDir:                  path.Join(dir, "cni-netd"),
			CNIPluginsBinary:            path.Join(dir, "mockcni"),
			CNIPlugins:                  []string{"plugin1", "plugin2"},
			UID:                         os.Getuid(),
			GID:                         os.Getgid(),
		},
	}

	// Ensure required directories are created and set up correctly.
	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())

	return s
}

func TestKubelet(t *testing.T) {
	g := NewWithT(t)

	s := testFixture(t)

	kubernetesPKIDir := s.KubernetesPKIDir()
	containerdSocketDir := s.ContainerdSocketDir()
	kubernetesConfigDir := s.KubernetesConfigDir()
	kubeletRootDir := s.KubeletRootDir()

	g.Expect(setup.KubeletControlPlane(s, "dev", nil, "10.152.1.1", "test-cluster.local", "")).To(BeNil())

	t.Run("Args", func(t *testing.T) {
		expectedLabels := strings.Join(append(setup.GetKubeletControlPlaneLabels(), setup.GetKubeletWorkerLabels()...), ",")
		tests := []struct {
			key         string
			expectedVal string
		}{
			{"--anonymous-auth", "false"},
			{"--authentication-token-webhook", "true"},
			{"--cert-dir", kubernetesPKIDir},
			{"--client-ca-file", path.Join(kubernetesPKIDir, "ca.crt")},
			{"--container-runtime-endpoint", path.Join(containerdSocketDir, "containerd.sock")},
			{"--containerd", path.Join(containerdSocketDir, "containerd.sock")},
			{"--eviction-hard", "'memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi'"},
			{"--fail-swap-on", "false"},
			{"--hostname-override", "dev"},
			{"--kubeconfig", path.Join(kubernetesConfigDir, "kubelet.conf")},
			{"--node-labels", expectedLabels},
			{"--read-only-port", "0"},
			{"--register-with-taints", ""},
			{"--root-dir", kubeletRootDir},
			{"--serialize-image-pulls", "false"},
			{"--tls-cipher-suites", strings.Join(setup.GetKubeletTLSCipherSuites(), ",")},
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
