package v1_test

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestBootstrapConfigToMicrocluster(t *testing.T) {
	g := NewWithT(t)

	cfg := apiv1.BootstrapConfig{
		ClusterConfig: apiv1.UserFacingClusterConfig{
			Network: apiv1.NetworkConfig{
				Enabled: vals.Pointer(true),
			},
			DNS: apiv1.DNSConfig{
				Enabled:       vals.Pointer(true),
				ClusterDomain: vals.Pointer("cluster.local"),
			},
			Ingress: apiv1.IngressConfig{
				Enabled: vals.Pointer(true),
			},
			LoadBalancer: apiv1.LoadBalancerConfig{
				Enabled: vals.Pointer(true),
				L2Mode:  vals.Pointer(true),
				CIDRs:   vals.Pointer([]string{"10.0.0.0/24", "10.1.0.10-10.1.0.20"}),
			},
			LocalStorage: apiv1.LocalStorageConfig{
				Enabled:   vals.Pointer(true),
				LocalPath: vals.Pointer("/storage/path"),
				Default:   vals.Pointer(false),
			},
			Gateway: apiv1.GatewayConfig{
				Enabled: vals.Pointer(true),
			},
			MetricsServer: apiv1.MetricsServerConfig{
				Enabled: vals.Pointer(true),
			},
			CloudProvider: vals.Pointer("external"),
		},
		PodCIDR:       vals.Pointer("10.100.0.0/16"),
		ServiceCIDR:   vals.Pointer("10.200.0.0/16"),
		DisableRBAC:   vals.Pointer(false),
		SecurePort:    vals.Pointer(6443),
		K8sDqlitePort: vals.Pointer(9090),
		DatastoreType: vals.Pointer("k8s-dqlite"),
		ExtraSANs:     []string{"custom.kubernetes"},
	}

	microclusterConfig, err := cfg.ToMicrocluster()
	g.Expect(err).To(BeNil())

	fromMicrocluster, err := apiv1.BootstrapConfigFromMicrocluster(microclusterConfig)
	g.Expect(err).To(BeNil())
	g.Expect(fromMicrocluster).To(Equal(cfg))
}
