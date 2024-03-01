package k8s

import (
	"os"
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
enable-rbac: true`

var yamlConfigInvalidYaml = `
components:
strawberries
  - apples
  - oranges`

func mustCreateTemporaryTestDirectory(t *testing.T) string {
	// Create a temporary test directory to mock the snap
	// <tempDir>
	// 	└── init.yaml
	t.Helper()

	tempDir := t.TempDir()

	err := os.MkdirAll(tempDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	return tempDir
}

func mustAddConfigToTestDir(t *testing.T, path string, data string) {
	t.Helper()
	// Create the init botstrap config file
	err := os.WriteFile(path+"/init.yaml", []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConfigYaml(t *testing.T) {
	t.Run("CompleteConfig", func(t *testing.T) {
		g := NewWithT(t)

		tempDir := mustCreateTemporaryTestDirectory(t)
		configPath := tempDir + "/init.yaml"

		// Add the complete config to the test directory
		mustAddConfigToTestDir(t, tempDir, yamlConfigComplete)

		// Get the config from the test directory
		bootstrapConfig, err := getConfigFromYaml(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
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

		tempDir := mustCreateTemporaryTestDirectory(t)
		configPath := tempDir + "/init.yaml"

		// Add the incomplete config to the test directory
		mustAddConfigToTestDir(t, tempDir, yamlConfigIncomplete)

		// Get the config from the test directory
		bootstrapConfig, err := getConfigFromYaml(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
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

		tempDir := mustCreateTemporaryTestDirectory(t)
		configPath := tempDir + "/init.yaml"

		// Add the invalid yaml to the test directory
		mustAddConfigToTestDir(t, tempDir, yamlConfigInvalidYaml)

		// Get the config from the test directory
		_, err := getConfigFromYaml(configPath)
		g.Expect(err).NotTo(BeNil())
	})

}
