package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func mockK8sDqliteEmbeddedSnap(t *testing.T) *mock.Snap {
	s := &mock.Snap{
		Mock: mock.Mock{
			ServiceArgumentsDir: t.TempDir(),
			K8sDqliteStateDir:   t.TempDir(),
			EtcdPKIDir:          t.TempDir(),
			KubernetesPKIDir:    t.TempDir(),
		},
	}

	NewWithT(t).Expect(setup.EnsureAllDirectories(s)).To(Succeed())
	return s
}

func TestK8sDqliteEmbedded(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockK8sDqliteEmbeddedSnap(t)

		// Call the K8sDqlite setup function with mock arguments
		g.Expect(setup.K8sDqliteEmbedded(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", nil, nil)).To(BeNil())

		// Ensure the K8sDqlite arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--embedded", expectedVal: "true"},
			{key: "--storage-dir", expectedVal: s.K8sDqliteStateDir()},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-dqlite", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		// Ensure the K8sDqlite arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(filepath.Join(s.ServiceArgumentsDir(), "k8s-dqlite"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("YAMLFileContents", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockK8sDqliteEmbeddedSnap(t)
		g.Expect(setup.K8sDqliteEmbedded(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", nil, nil)).To(BeNil())

		eb, err := os.ReadFile(filepath.Join(s.K8sDqliteStateDir(), "embedded.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(eb)).To(SatisfyAll(
			ContainSubstring("initial-cluster-state: new"),
			ContainSubstring("data-dir: %s/data", s.K8sDqliteStateDir()),
			ContainSubstring("name: t1"),
			ContainSubstring("advertise-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster-state: new"),
			ContainSubstring("initial-advertise-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster: t1=https://127.0.0.1:2380"),
		))

		cb, err := os.ReadFile(filepath.Join(s.K8sDqliteStateDir(), "config.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(cb)).To(SatisfyAll(
			ContainSubstring("peer-url: https://127.0.0.1:2380"),
			ContainSubstring("peer-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("peer-cert-file: %s/peer.crt", s.EtcdPKIDir()),
			ContainSubstring("peer-key-file: %s/peer.key", s.EtcdPKIDir()),
			ContainSubstring("ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("cert-file: %s/server.crt", s.EtcdPKIDir()),
			ContainSubstring("key-file: %s/server.key", s.EtcdPKIDir()),
		))
	})

	t.Run("JoiningNode", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockK8sDqliteEmbeddedSnap(t)
		g.Expect(setup.K8sDqliteEmbedded(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", []string{"https://10.0.0.1:2379"}, nil)).To(BeNil())

		eb, err := os.ReadFile(filepath.Join(s.K8sDqliteStateDir(), "embedded.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(eb)).To(SatisfyAll(
			ContainSubstring("data-dir: %s/data", s.K8sDqliteStateDir()),
			ContainSubstring("name: t1"),
			ContainSubstring("advertise-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster-state: existing"),
			ContainSubstring("initial-advertise-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster: t1=https://127.0.0.1:2380"),
		))

		cb, err := os.ReadFile(filepath.Join(s.K8sDqliteStateDir(), "config.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(cb)).To(SatisfyAll(
			ContainSubstring("client-urls:\n- https://10.0.0.1:2379"),
			ContainSubstring("peer-url: https://127.0.0.1:2380"),
			ContainSubstring("peer-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("peer-cert-file: %s/peer.crt", s.EtcdPKIDir()),
			ContainSubstring("peer-key-file: %s/peer.key", s.EtcdPKIDir()),
			ContainSubstring("ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("cert-file: %s/server.crt", s.EtcdPKIDir()),
			ContainSubstring("key-file: %s/server.key", s.EtcdPKIDir()),
		))
	})

	t.Run("MissingStateDir", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockK8sDqliteEmbeddedSnap(t)
		s.Mock.K8sDqliteStateDir = "nonexistent"
		g.Expect(setup.K8sDqlite(s, "", []string{}, nil)).ToNot(Succeed())
	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockK8sDqliteEmbeddedSnap(t)
		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.K8sDqlite(s, "", []string{}, nil)).ToNot(Succeed())
	})
}
