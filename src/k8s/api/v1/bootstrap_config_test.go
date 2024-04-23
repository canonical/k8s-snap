package v1_test

import (
	"github.com/canonical/k8s/pkg/utils"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	. "github.com/onsi/gomega"
)

func TestBootstrapConfigToMicrocluster(t *testing.T) {
	g := NewWithT(t)

	cfg := apiv1.BootstrapConfig{
		ClusterConfig: apiv1.UserFacingClusterConfig{
			Network: apiv1.NetworkConfig{
				Enabled: utils.Pointer(true),
			},
			DNS: apiv1.DNSConfig{
				Enabled:       utils.Pointer(true),
				ClusterDomain: utils.Pointer("cluster.local"),
			},
			Ingress: apiv1.IngressConfig{
				Enabled: utils.Pointer(true),
			},
			LoadBalancer: apiv1.LoadBalancerConfig{
				Enabled: utils.Pointer(true),
				L2Mode:  utils.Pointer(true),
				CIDRs:   utils.Pointer([]string{"10.0.0.0/24", "10.1.0.10-10.1.0.20"}),
			},
			LocalStorage: apiv1.LocalStorageConfig{
				Enabled:   utils.Pointer(true),
				LocalPath: utils.Pointer("/storage/path"),
				Default:   utils.Pointer(false),
			},
			Gateway: apiv1.GatewayConfig{
				Enabled: utils.Pointer(true),
			},
			MetricsServer: apiv1.MetricsServerConfig{
				Enabled: utils.Pointer(true),
			},
			CloudProvider: utils.Pointer("external"),
		},
		PodCIDR:       utils.Pointer("10.100.0.0/16"),
		ServiceCIDR:   utils.Pointer("10.200.0.0/16"),
		DisableRBAC:   utils.Pointer(false),
		SecurePort:    utils.Pointer(6443),
		K8sDqlitePort: utils.Pointer(9090),
		DatastoreType: utils.Pointer("k8s-dqlite"),
		ExtraSANs:     []string{"custom.kubernetes"},
	}

	microclusterConfig, err := cfg.ToMicrocluster()
	g.Expect(err).To(BeNil())

	fromMicrocluster, err := apiv1.BootstrapConfigFromMicrocluster(microclusterConfig)
	g.Expect(err).To(BeNil())
	g.Expect(fromMicrocluster).To(Equal(cfg))
}
