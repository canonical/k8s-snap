package newtypes_test

import (
	"testing"

	newtypes "github.com/canonical/k8s/pkg/k8sd/new"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestSetDefaults(t *testing.T) {
	g := NewWithT(t)
	clusterConfig := newtypes.ClusterConfig{}

	// Set defaults
	expectedConfig := newtypes.ClusterConfig{
		Network: newtypes.Network{
			PodCIDR:     vals.Pointer("10.1.0.0/16"),
			ServiceCIDR: vals.Pointer("10.152.183.0/24"),
		},
		APIServer: newtypes.APIServer{
			SecurePort:        vals.Pointer(6443),
			AuthorizationMode: vals.Pointer("Node,RBAC"),
		},
		Datastore: newtypes.Datastore{
			K8sDqlitePort: vals.Pointer(9000),
		},
		Features: newtypes.Features{
			DNS: newtypes.DNSFeature{
				UpstreamNameservers: vals.Pointer([]string{"/etc/resolv.conf"}),
			},
			LocalStorage: newtypes.LocalStorageFeature{
				LocalPath:     vals.Pointer("/var/snap/k8s/common/rawfile-storage"),
				ReclaimPolicy: vals.Pointer("Delete"),
				SetDefault:    vals.Pointer(true),
			},
			LoadBalancer: newtypes.LoadBalancerFeature{
				L2Mode: vals.Pointer(true),
			},
		},
	}

	clusterConfig.SetDefaults()
	g.Expect(clusterConfig).To(Equal(expectedConfig))
}
