package k8s

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

var (
	//go:embed testdata/bootstrap-config-full.yaml
	bootstrapConfigFull string
	//go:embed testdata/bootstrap-config-some.yaml
	bootstrapConfigSome string
	//go:embed testdata/bootstrap-config-invalid-keys.yaml
	bootstrapConfigInvalidKeys string
)

type testCase struct {
	name           string
	yamlConfig     string
	expectedConfig apiv1.BootstrapConfig
	expectedError  string
}

var testCases = []testCase{
	{
		name:       "FullConfig",
		yamlConfig: bootstrapConfigFull,
		expectedConfig: apiv1.BootstrapConfig{
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
					CIDRs:   vals.Pointer([]string{"10.0.0.0/24"}),
				},
				LocalStorage: apiv1.LocalStorageConfig{
					Enabled:    vals.Pointer(true),
					LocalPath:  vals.Pointer("/storage/path"),
					SetDefault: vals.Pointer(false),
				},
				Gateway: apiv1.GatewayConfig{
					Enabled: vals.Pointer(true),
				},
				MetricsServer: apiv1.MetricsServerConfig{
					Enabled: vals.Pointer(true),
				},
			},
			PodCIDR:       vals.Pointer("10.100.0.0/16"),
			ServiceCIDR:   vals.Pointer("10.200.0.0/16"),
			DisableRBAC:   vals.Pointer(false),
			SecurePort:    vals.Pointer(6443),
			CloudProvider: vals.Pointer("external"),
			K8sDqlitePort: vals.Pointer(9090),
			DatastoreType: vals.Pointer("k8s-dqlite"),
			ExtraSANs:     []string{"custom.kubernetes"},
		},
	},
	{
		name:       "SomeConfig",
		yamlConfig: bootstrapConfigSome,
		expectedConfig: apiv1.BootstrapConfig{
			PodCIDR:     vals.Pointer("10.100.0.0/16"),
			ServiceCIDR: vals.Pointer("10.152.200.0/24"),
		},
	},
	{
		name:          "InvalidKeys",
		yamlConfig:    bootstrapConfigInvalidKeys,
		expectedError: "field cluster-cidr not found in type v1.BootstrapConfig",
	},
	{
		name:          "InvalidYAML",
		yamlConfig:    "this is not valid yaml",
		expectedError: "failed to parse YAML config file",
	},
}

func mustAddConfigToTestDir(t *testing.T, configPath string, data string) {
	t.Helper()
	// Create the cluster bootstrap config file
	err := os.WriteFile(configPath, []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConfigYaml(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "init.yaml")

			// Add the test case config to the test directory
			mustAddConfigToTestDir(t, configPath, tc.yamlConfig)

			// Get the config from the test directory
			bootstrapConfig, err := getConfigFromYaml(configPath)

			if tc.expectedError != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tc.expectedError))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(bootstrapConfig).To(Equal(tc.expectedConfig))
			}
		})
	}
}
