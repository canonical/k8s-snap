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
	t.Run("NewNode", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)

		// Call the Etcd setup function with mock arguments
		g.Expect(setup.Etcd(s, "t1", net.ParseIP("10.0.0.3"), 2379, 2380, nil, nil)).To(Succeed())

		// Ensure the Etcd arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--data-dir", expectedVal: filepath.Join(s.EtcdDir(), "data")},
			{key: "--name", expectedVal: "t1"},
			{key: "--initial-advertise-peer-urls", expectedVal: "https://10.0.0.3:2380"},
			{key: "--listen-peer-urls", expectedVal: "https://10.0.0.3:2380"},
			{key: "--listen-client-urls", expectedVal: "https://10.0.0.3:2379,https://127.0.0.1:2379"},
			{key: "--advertise-client-urls", expectedVal: "https://10.0.0.3:2379"},
			{key: "--initial-cluster-state", expectedVal: "new"},
			{key: "--initial-cluster", expectedVal: "t1=https://10.0.0.3:2380"},
			{key: "--client-cert-auth", expectedVal: "true"},
			{key: "--trusted-ca-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "ca.crt")},
			{key: "--cert-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "server.crt")},
			{key: "--key-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "server.key")},
			{key: "--peer-client-cert-auth", expectedVal: "true"},
			{key: "--peer-trusted-ca-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "ca.crt")},
			{key: "--peer-cert-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "peer.crt")},
			{key: "--peer-key-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "peer.key")},
			{key: "--auto-tls", expectedVal: "false"},
			{key: "--peer-auto-tls", expectedVal: "false"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "etcd", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		args, err := utils.ParseArgumentFile(filepath.Join(s.ServiceArgumentsDir(), "etcd"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("JoiningNode", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		g.Expect(setup.Etcd(s, "t1", net.ParseIP("10.0.0.3"), 2379, 2380, map[string]string{"t2": "https://10.0.0.1:2380"}, nil)).To(Succeed())

		// Ensure the Etcd arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--data-dir", expectedVal: filepath.Join(s.EtcdDir(), "data")},
			{key: "--name", expectedVal: "t1"},
			{key: "--initial-advertise-peer-urls", expectedVal: "https://10.0.0.3:2380"},
			{key: "--listen-peer-urls", expectedVal: "https://10.0.0.3:2380"},
			{key: "--listen-client-urls", expectedVal: "https://10.0.0.3:2379,https://127.0.0.1:2379"},
			{key: "--advertise-client-urls", expectedVal: "https://10.0.0.3:2379"},
			{key: "--initial-cluster-state", expectedVal: "existing"},
			{key: "--initial-cluster", expectedVal: "t1=https://10.0.0.3:2380,t2=https://10.0.0.1:2380"},
			{key: "--client-cert-auth", expectedVal: "true"},
			{key: "--trusted-ca-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "ca.crt")},
			{key: "--cert-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "server.crt")},
			{key: "--key-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "server.key")},
			{key: "--peer-client-cert-auth", expectedVal: "true"},
			{key: "--peer-trusted-ca-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "ca.crt")},
			{key: "--peer-cert-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "peer.crt")},
			{key: "--peer-key-file", expectedVal: filepath.Join(s.EtcdPKIDir(), "peer.key")},
			{key: "--auto-tls", expectedVal: "false"},
			{key: "--peer-auto-tls", expectedVal: "false"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "etcd", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		args, err := utils.ParseArgumentFile(filepath.Join(s.ServiceArgumentsDir(), "etcd"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))
	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mockEtcdSnap(t)
		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.Etcd(s, "", nil, 0, 0, nil, nil)).ToNot(Succeed())
	})
}
