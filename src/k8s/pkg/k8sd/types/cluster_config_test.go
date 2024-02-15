package types_test

import (
	"fmt"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestSetClusterConfigDefaults(t *testing.T) {
	g := NewWithT(t)
	bootstrapConfig := apiv1.BootstrapConfig{
		ClusterCIDR: "10.100.0.0/16",
		Components:  []string{"dns", "network"},
	}

	expectedConfig := types.ClusterConfig{
		Network: types.Network{
			PodCIDR:     "10.100.0.0/16", // not default pod CIDR
			ServiceCIDR: "10.152.183.0/24",
		},
		APIServer: types.APIServer{
			Datastore:         "k8s-dqlite",
			SecurePort:        6443,
			AuthorizationMode: "Node,RBAC",
		},
		K8sDqlite: types.K8sDqlite{
			Port: 9000,
		},
	}

	conf, err := types.SetClusterConfigDefaults(&bootstrapConfig)

	g.Expect(err).To(BeNil())
	g.Expect(conf).To(Equal(expectedConfig))

	// Test invalid cluster CIDR
	invalidBootstrapConfig := apiv1.BootstrapConfig{
		ClusterCIDR: "Bananas.0.0/16",
		Components:  []string{"dns", "network"},
	}

	conf, err = types.SetClusterConfigDefaults(&invalidBootstrapConfig)
	fmt.Println(err)
	g.Expect(err).ToNot(BeNil())
	g.Expect(conf).To(Equal(types.ClusterConfig{}))

}

type mergeClusterConfigTestCase struct {
	name         string
	old          types.ClusterConfig
	new          types.ClusterConfig
	expectResult types.ClusterConfig
	expectErr    bool
}

func generateMergeClusterConfigTestCases(field string, changeAllowed bool, val1 any, val2 any, update func(*types.ClusterConfig, any)) []mergeClusterConfigTestCase {
	var cfgZero, cfgOne, cfgTwo types.ClusterConfig
	update(&cfgOne, val1)
	update(&cfgTwo, val2)

	return []mergeClusterConfigTestCase{
		{
			name:         fmt.Sprintf("%s/Set", field),
			new:          cfgOne,
			expectResult: cfgOne,
			expectErr:    false,
		},
		{
			name:         fmt.Sprintf("%s/Keep", field),
			old:          cfgOne,
			new:          cfgZero,
			expectResult: cfgOne,
		},
		{
			name:         fmt.Sprintf("%s/Update", field),
			old:          cfgOne,
			new:          cfgTwo,
			expectResult: cfgTwo,
			expectErr:    !changeAllowed,
		},
	}
}

func TestMergeClusterConfig(t *testing.T) {
	for _, tcs := range [][]mergeClusterConfigTestCase{
		generateMergeClusterConfigTestCases("CACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CACert = v.(string) }),
		generateMergeClusterConfigTestCases("CAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CAKey = v.(string) }),
		generateMergeClusterConfigTestCases("K8sDqliteCert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.K8sDqliteCert = v.(string) }),
		generateMergeClusterConfigTestCases("K8sDqliteKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.K8sDqliteKey = v.(string) }),
		generateMergeClusterConfigTestCases("APIServerKubeletClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.APIServerKubeletClientCert = v.(string) }),
		generateMergeClusterConfigTestCases("APIServerKubeletClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.APIServerKubeletClientKey = v.(string) }),
		generateMergeClusterConfigTestCases("FrontProxyCACert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCACert = v.(string) }),
		generateMergeClusterConfigTestCases("FrontProxyCAKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCAKey = v.(string) }),
		generateMergeClusterConfigTestCases("AuthorizationMode", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.AuthorizationMode = v.(string) }),
		generateMergeClusterConfigTestCases("ServiceAccountKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.ServiceAccountKey = v.(string) }),
		generateMergeClusterConfigTestCases("PodCIDR", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Network.PodCIDR = v.(string) }),
		generateMergeClusterConfigTestCases("ServiceCIDR", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Network.ServiceCIDR = v.(string) }),
		generateMergeClusterConfigTestCases("Datastore", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.Datastore = v.(string) }),
		generateMergeClusterConfigTestCases("DatastoreURL", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.DatastoreURL = v.(string) }),
		generateMergeClusterConfigTestCases("DatastoreCA", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.DatastoreCA = v.(string) }),
		generateMergeClusterConfigTestCases("DatastoreClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.DatastoreClientCert = v.(string) }),
		generateMergeClusterConfigTestCases("DatastoreClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.DatastoreClientKey = v.(string) }),
		generateMergeClusterConfigTestCases("ClusterDNS", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDNS = v.(string) }),
		generateMergeClusterConfigTestCases("ClusterDomain", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDomain = v.(string) }),
		generateMergeClusterConfigTestCases("CloudProvider", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.CloudProvider = v.(string) }),
		generateMergeClusterConfigTestCases("SecurePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.APIServer.SecurePort = v.(int) }),
		generateMergeClusterConfigTestCases("K8sDqlitePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.K8sDqlite.Port = v.(int) }),
	} {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				result, err := types.MergeClusterConfig(tc.old, tc.new)
				if tc.expectErr {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(result).To(Equal(tc.expectResult))
				}
			})
		}
	}
}
