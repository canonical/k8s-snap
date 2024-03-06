package k8s

import (
	"os"
	"path/filepath"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

type testCase struct {
	name           string
	yamlConfig     string
	expectedConfig apiv1.BootstrapConfig
	expectedError  string
}

var testCases = []testCase{
	{
		name: "CompleteConfig",
		yamlConfig: `
components:
  - network
  - dns
  - gateway
  - ingress
  - storage
  - metrics-server
cluster-cidr: "10.244.0.0/16"
service-cidr: "10.152.100.0/24"
enable-rbac: true
k8s-dqlite-port: 12379`,
		expectedConfig: apiv1.BootstrapConfig{
			Components:    []string{"network", "dns", "gateway", "ingress", "storage", "metrics-server"},
			ClusterCIDR:   "10.244.0.0/16",
			ServiceCIDR:   "10.152.100.0/24",
			EnableRBAC:    vals.Pointer(true),
			K8sDqlitePort: 12379,
		},
	},
	{
		name: "IncompleteConfig",
		yamlConfig: `
cluster-cidr: "10.244.0.0/16"
enable-rbac: true
bananas: 5`,
		expectedConfig: apiv1.BootstrapConfig{
			Components:    []string{"dns", "metrics-server", "network"},
			ClusterCIDR:   "10.244.0.0/16",
			ServiceCIDR:   "10.152.183.0/24",
			EnableRBAC:    vals.Pointer(true),
			K8sDqlitePort: 9000,
		},
	},
	{
		name:          "InvalidYaml",
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
