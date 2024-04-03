package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestDatastoreToKubeAPIServerArguments(t *testing.T) {
	snap := &mock.Snap{
		Mock: mock.Mock{
			K8sDqliteStateDir: "/k8s-dqlite",
			EtcdPKIDir:        "/pki/etcd",
		},
	}

	for _, tc := range []struct {
		name             string
		config           types.Datastore
		expectUpdateArgs map[string]string
		expectDeleteArgs []string
	}{
		{
			name:             "Nil",
			expectUpdateArgs: map[string]string{},
		},
		{
			name: "K8sDqlite",
			config: types.Datastore{
				Type: vals.Pointer("k8s-dqlite"),
			},
			expectUpdateArgs: map[string]string{
				"--etcd-servers": "unix:///k8s-dqlite/k8s-dqlite.sock",
			},
			expectDeleteArgs: []string{"--etcd-cafile", "--etcd-certfile", "--etcd-keyfile"},
		},
		{
			name: "ExternalFull",
			config: types.Datastore{
				Type:               vals.Pointer("external"),
				ExternalURL:        vals.Pointer("https://10.0.0.10:2379,https://10.0.0.11:2379"),
				ExternalCACert:     vals.Pointer("data"),
				ExternalClientCert: vals.Pointer("data"),
				ExternalClientKey:  vals.Pointer("data"),
			},
			expectUpdateArgs: map[string]string{
				"--etcd-servers":  "https://10.0.0.10:2379,https://10.0.0.11:2379",
				"--etcd-cafile":   "/pki/etcd/ca.crt",
				"--etcd-certfile": "/pki/etcd/client.crt",
				"--etcd-keyfile":  "/pki/etcd/client.key",
			},
		},
		{
			name: "ExternalOnlyCA",
			config: types.Datastore{
				Type:           vals.Pointer("external"),
				ExternalURL:    vals.Pointer("https://10.0.0.10:2379,https://10.0.0.11:2379"),
				ExternalCACert: vals.Pointer("data"),
			},
			expectUpdateArgs: map[string]string{
				"--etcd-servers": "https://10.0.0.10:2379,https://10.0.0.11:2379",
				"--etcd-cafile":  "/pki/etcd/ca.crt",
			},
			expectDeleteArgs: []string{"--etcd-certfile", "--etcd-keyfile"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			update, delete := tc.config.ToKubeAPIServerArguments(snap)
			g.Expect(update).To(Equal(tc.expectUpdateArgs))
			g.Expect(delete).To(Equal(tc.expectDeleteArgs))
		})
	}
}
