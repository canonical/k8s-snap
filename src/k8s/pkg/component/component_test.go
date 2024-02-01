package component

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/time"
)

var verbose = flag.Bool("test.log", false, "enable test logging")

func actionConfigFixture(t *testing.T) *action.Configuration {
	t.Helper()

	registryClient, err := registry.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	return &action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			if *verbose {
				t.Logf(format, v...)
			}
		},
	}
}

type chartOptions struct {
	*chart.Chart
}

type chartOption func(*chartOptions)

func buildChart(opts ...chartOption) *chart.Chart {
	c := &chartOptions{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v1",
				Name:       "hello",
				Version:    "0.1.0",
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c.Chart
}

func namedReleaseStub(name string, status release.Status) *release.Release {
	now := time.Now()
	return &release.Release{
		Name: name,
		Info: &release.Info{
			FirstDeployed: now,
			LastDeployed:  now,
			Status:        status,
			Description:   "Named Release Stub",
		},
		Config:  map[string]interface{}{"name": "value"},
		Version: 1,
	}
}

// "Mock" Initializer
type MockHelmClientInitializer struct {
	actionConfig *action.Configuration
}

func (r *MockHelmClientInitializer) InitializeHelmClientConfig() (*action.Configuration, error) {
	return r.actionConfig, nil
}

func makeMeSomeReleases(store *storage.Storage, t *testing.T) {
	t.Helper()
	relStub1 := namedReleaseStub("whiskas-1", release.StatusDeployed)
	relStub2 := namedReleaseStub("whiskas-2", release.StatusDeployed)
	relStub3 := namedReleaseStub("whiskas-3", release.StatusDeployed)

	for _, rel := range []*release.Release{relStub1, relStub2, relStub3} {
		if err := store.Create(rel); err != nil {
			t.Fatal(err)
		}
	}

	all, err := store.ListReleases()
	assert.NoError(t, err)
	assert.Len(t, all, 3, "sanity test: three items added")
}

var componentsNone = ``

var components = `
one:
  release: "whiskas-1"
  chart: "chunky-tuna-1.14.1.tgz"
  namespace: "default"
two:
  release: "whiskas-2"
  chart: "tuna-1.29.0.tgz"
  namespace: "default"
three:
  release: "whiskas-3"
  chart: "chunky-1.29.0.tgz"
  namespace: "default"
`

func createTemporaryTestDirectory(t *testing.T) string {
	// Create a temporary test directory to mock the snap
	// <tempDir>
	// └── k8s/components
	// 	├── charts
	// 	└── component.yaml
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}

	k8sComponentsDir := filepath.Join(tempDir, "k8s", "components")
	err = os.MkdirAll(k8sComponentsDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	k8sComponentsChartsDir := filepath.Join(k8sComponentsDir, "charts")
	err = os.MkdirAll(k8sComponentsChartsDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	return tempDir
}

func addConfigToTestDir(t *testing.T, path string, data string) {
	// Create a file and add some configs
	err := ioutil.WriteFile(path, []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func createNewManager(t *testing.T, components string) (*helmClient, string, *action.Configuration) {
	// Create a mock actionConfig for testing
	mockActionConfig := actionConfigFixture(t)
	// Create a mock HelmClient with the desired behavior for testing
	mockClient := &MockHelmClientInitializer{actionConfig: mockActionConfig}

	// create test directory to use for the snap mock
	tempDir := createTemporaryTestDirectory(t)

	// Create a file and add some configs
	addConfigToTestDir(t, filepath.Join(tempDir, "k8s", "components", "components.yaml"), components)

	// Create mock snap
	snap := &mock.Snap{
		PathPrefix: tempDir,
	}

	//Create a mock ComponentManager with the mock HelmClient
	mockComponentManager, err := NewManager(snap, mockClient)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, mockComponentManager)
	assert.IsType(t, &helmClient{}, mockComponentManager)
	return mockComponentManager, tempDir, mockActionConfig
}

func TestNewManager(t *testing.T) {
	// Create a mock actionConfig for testing
	mockHelmClient, tempDir, mockActionConfig := createNewManager(t, components)
	defer os.RemoveAll(tempDir)

	assert.NotNil(t, mockHelmClient)
	assert.IsType(t, &helmClient{}, mockHelmClient)
	assert.IsType(t, &MockHelmClientInitializer{}, mockHelmClient.initializer)
	assert.IsType(t, &mock.Snap{}, mockHelmClient.snap)
	assert.IsType(t, &action.Configuration{}, mockActionConfig)
	assert.DirExists(t, tempDir)
}

func TestListEmpty(t *testing.T) {
	// Create a mock ComponentManager with no components
	mockHelmClient, tempDir, _ := createNewManager(t, componentsNone)
	defer os.RemoveAll(tempDir)

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(components))
	}
}

func TestList(t *testing.T) {
	// Create a mock ComponentManager with the mock HelmClient
	// This mock uses components.yaml for the snap mock components

	mockHelmClient, tempDir, mockActionConfig := createNewManager(t, components)
	defer os.RemoveAll(tempDir)

	// Create releases in the mock actionConfig
	makeMeSomeReleases(mockActionConfig.Releases, t)

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, components)
	assert.Equal(t, 3, len(components))

	assert.Contains(t, components, Component{Name: "one", Status: true})
	assert.Contains(t, components, Component{Name: "two", Status: true})
	assert.Contains(t, components, Component{Name: "three", Status: true})

}
