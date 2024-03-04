package k8s

import (
	"os"
	"path/filepath"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	. "github.com/onsi/gomega"
)

var yamlConfigComplete = `
components:
  - network
  - dns
  - gateway
  - ingress
  - storage
  - metrics-server
cluster-cidr: "10.244.0.0/16"
enable-rbac: true
k8s-dqlite-port: 12379`

var yamlConfigIncomplete = `
cluster-cidr: "10.244.0.0/16"
enable-rbac: true
bananas: 5`

func mustAddConfigToTestDir(t *testing.T, configPath string, data string) {
	t.Helper()
	// Create the cluster bootstrap config file
	err := os.WriteFile(configPath, []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConfigYaml(t *testing.T) {
	t.Run("CompleteConfig", func(t *testing.T) {
		g := NewWithT(t)

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "init.yaml")

		// Add the complete config to the test directory
		mustAddConfigToTestDir(t, configPath, yamlConfigComplete)

		// Get the config from the test directory
		bootstrapConfig, err := getConfigFromYaml(configPath)
		if err != nil {
			t.Fatalf("failed to load bootstrap configuration file: %v", err)
		}

		// Check the config
		expectedConfig := apiv1.BootstrapConfig{
			Components:    []string{"network", "dns", "gateway", "ingress", "storage", "metrics-server"},
			ClusterCIDR:   "10.244.0.0/16",
			EnableRBAC:    &[]bool{true}[0],
			K8sDqlitePort: 12379,
		}
		g.Expect(bootstrapConfig).To(Equal(expectedConfig))

	})
	t.Run("IncompleteConfig", func(t *testing.T) {
		// test an incomplete config file, set defaults for unspecified fields
		g := NewWithT(t)

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "init.yaml")

		// Add the complete config to the test directory
		mustAddConfigToTestDir(t, configPath, yamlConfigIncomplete)

		// Get the config from the test directory
		bootstrapConfig, err := getConfigFromYaml(configPath)
		if err != nil {
			t.Fatalf("failed to load bootstrap configuration file: %v", err)
		}

		// Check the config
		expectedConfig := apiv1.BootstrapConfig{
			Components:    []string{"dns", "metrics-server", "network"},
			ClusterCIDR:   "10.244.0.0/16",
			EnableRBAC:    &[]bool{true}[0],
			K8sDqlitePort: 9000,
		}
		g.Expect(bootstrapConfig).To(Equal(expectedConfig))
	})

	t.Run("InvalidYaml", func(t *testing.T) {
		// test an invalid yaml file
		g := NewWithT(t)

		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "init.yaml")

		// Add the invalid yaml to the test directory
		mustAddConfigToTestDir(t, configPath, "this is not valid yaml")

		// Get the config from the test directory
		_, err := getConfigFromYaml(configPath)
		g.Expect(err.Error()).To(ContainSubstring("failed to parse YAML config file"))
	})

}
