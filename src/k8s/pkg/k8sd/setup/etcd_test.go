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

func mockEtcdSnap(t *testing.T) *mock.Snap {
	s := &mock.Snap{
		Mock: mock.Mock{
			ServiceArgumentsDir: t.TempDir(),
			EtcdDir:             t.TempDir(),
			EtcdPKIDir:          t.TempDir(),
			KubernetesPKIDir:    t.TempDir(),
		},
	}

	NewWithT(t).Expect(setup.EnsureAllDirectories(s)).To(Succeed())
	return s
}

func TestEtcd(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)

		// Call the Etcd setup function with mock arguments
		g.Expect(setup.Etcd(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", nil, nil)).To(BeNil())

		// Ensure the K8sDqlite arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--etcd-mode", expectedVal: "true"},
			{key: "--storage-dir", expectedVal: s.EtcdDir()},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-dqlite", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		args, err := utils.ParseArgumentFile(filepath.Join(s.ServiceArgumentsDir(), "k8s-dqlite"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("YAMLFileContents", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		g.Expect(setup.Etcd(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", nil, nil)).To(BeNil())

		eb, err := os.ReadFile(filepath.Join(s.EtcdDir(), "etcd.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(eb)).To(SatisfyAll(
			ContainSubstring("initial-cluster-state: new"),
			ContainSubstring("data-dir: %s/data", s.EtcdDir()),
			ContainSubstring("name: t1"),
			ContainSubstring("advertise-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster-state: new"),
			ContainSubstring("initial-advertise-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster: t1=https://127.0.0.1:2380"),
			ContainSubstring("client-transport-security:"),
			ContainSubstring("  trusted-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("  cert-file: %s/server.crt", s.EtcdPKIDir()),
			ContainSubstring("  key-file: %s/server.key", s.EtcdPKIDir()),
			ContainSubstring("peer-transport-security:"),
			ContainSubstring("  trusted-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("  cert-file: %s/peer.crt", s.EtcdPKIDir()),
			ContainSubstring("  key-file: %s/peer.key", s.EtcdPKIDir()),
		))

		cb, err := os.ReadFile(filepath.Join(s.EtcdDir(), "register.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(cb)).To(SatisfyAll(
			ContainSubstring("peer-url: https://127.0.0.1:2380"),
			ContainSubstring("trusted-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("cert-file: %s/server.crt", s.EtcdPKIDir()),
			ContainSubstring("key-file: %s/server.key", s.EtcdPKIDir()),
		))
	})

	t.Run("JoiningNode", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		g.Expect(setup.Etcd(s, "t1", "https://127.0.0.1:2379", "https://127.0.0.1:2380", []string{"https://10.0.0.1:2379"}, nil)).To(BeNil())

		eb, err := os.ReadFile(filepath.Join(s.EtcdDir(), "etcd.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(eb)).To(SatisfyAll(
			ContainSubstring("data-dir: %s/data", s.EtcdDir()),
			ContainSubstring("name: t1"),
			ContainSubstring("advertise-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-client-urls: https://127.0.0.1:2379"),
			ContainSubstring("listen-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster-state: existing"),
			ContainSubstring("initial-advertise-peer-urls: https://127.0.0.1:2380"),
			ContainSubstring("initial-cluster: t1=https://127.0.0.1:2380"),
		))

		cb, err := os.ReadFile(filepath.Join(s.EtcdDir(), "register.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(cb)).To(SatisfyAll(
			ContainSubstring("client-urls:\n- https://10.0.0.1:2379"),
			ContainSubstring("peer-url: https://127.0.0.1:2380"),
			ContainSubstring("trusted-ca-file: %s/ca.crt", s.EtcdPKIDir()),
			ContainSubstring("cert-file: %s/server.crt", s.EtcdPKIDir()),
			ContainSubstring("key-file: %s/server.key", s.EtcdPKIDir()),
		))
	})

	t.Run("MissingStateDir", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		s.Mock.EtcdDir = "nonexistent"
		g.Expect(setup.Etcd(s, "", "", "", nil, nil)).ToNot(Succeed())
	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.Etcd(s, "", "", "", nil, nil)).ToNot(Succeed())
	})
}
