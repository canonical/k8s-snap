package component

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/component/mock"
	componentmock "github.com/canonical/k8s/pkg/component/mock"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/time"
)

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
	}
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

func mustMakeMeSomeReleases(store *storage.Storage, t *testing.T) (all []*release.Release) {
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
	if err != nil {
		t.Fatal(err)
	}

	return all
}

var componentsNone = ``

func mustCreateTemporaryTestDirectory(t *testing.T) string {
	// Create a temporary test directory to mock the snap
	// <tempDir>
	// └── k8s/components
	// 	├── charts
	// 	└── component.yaml
	tempDir := t.TempDir()

	k8sComponentsDir := filepath.Join(tempDir, "k8s", "components", "charts")
	err := os.MkdirAll(k8sComponentsDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	return tempDir
}

func mustAddConfigToTestDir(t *testing.T, path string, data string) {
	// Create a file and add some configs
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func mustCreateNewHelmClient(t *testing.T, components map[string]types.Component) (*helmClient, string, *action.Configuration) {
	// Create a mock actionConfig for testing
	mockActionConfig := actionConfigFixture(t)
	// Create a mock HelmClient with the desired behavior for testing
	mockClient := &componentmock.MockHelmConfigProvider{ActionConfig: mockActionConfig}

	// create test directory to use for the snap mock
	tempDir := mustCreateTemporaryTestDirectory(t)

	// Create mock snap
	snap := &snapmock.Snap{
		Mock: snapmock.Mock{
			Components: components,
		},
	}

	mockLoader := &mock.ChartLoader{}
	mockInstaller := &mock.ChartInstaller{}

	//Create a mock ComponentManager with the mock HelmClient
	mockHelmCLient, err := NewHelmClient(snap, mockClient, WithChartLoader(mockLoader), WithChartInstaller(mockInstaller))
	if err != nil {
		t.Fatal(err)
	}

	return mockHelmCLient, tempDir, mockActionConfig
}

func TestListEmptyComponents(t *testing.T) {
	g := NewWithT(t)
	// Create a mock ComponentManager with no components
	mockHelmClient, tempDir, _ := mustCreateNewHelmClient(t, nil)
	defer os.RemoveAll(tempDir)

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()

	g.Expect(err).To(BeNil())
	g.Expect(components).To(HaveLen(0))
}

func TestListComponentsWithReleases(t *testing.T) {
	g := NewWithT(t)

	// Create a mock ComponentManager with the mock HelmClient
	// This mock uses components.yaml for the snap mock components
	mockHelmClient, tempDir, mockActionConfig := mustCreateNewHelmClient(t, map[string]types.Component{
		"one": {
			ReleaseName:  "whiskas-1",
			Namespace:    "default",
			ManifestPath: "chunky-tuna-1.14.1.tgz",
		},
		"two": {
			ReleaseName:  "whiskas-2",
			Namespace:    "default",
			ManifestPath: "tuna-1.29.0.tgz",
		},
		"three": {
			ReleaseName:  "whiskas-3",
			Namespace:    "default",
			ManifestPath: "chunky-1.29.0.tgz",
		},
	})

	defer os.RemoveAll(tempDir)

	// Create releases in the mock actionConfig
	releases := mustMakeMeSomeReleases(mockActionConfig.Releases, t)
	g.Expect(releases).To(HaveLen(3))

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()

	g.Expect(err).To(BeNil())
	g.Expect(components).To(Equal([]Component{
		{Name: "one", Status: true},
		{Name: "three", Status: true},
		{Name: "two", Status: true},
	}))
}

func TestComponentEnableDisable(t *testing.T) {
	g := NewWithT(t)

	// Setup the mock environment
	mockHelmClient, tempDir, mockActionConfig := mustCreateNewHelmClient(t, map[string]types.Component{
		"one": {
			ReleaseName:  "whiskas-1",
			Namespace:    "default",
			ManifestPath: "chunky-tuna-1.14.1.tgz",
		},
	})

	os.RemoveAll(tempDir)

	releases := mustMakeMeSomeReleases(mockActionConfig.Releases, t)
	g.Expect(releases).To(HaveLen(3))

	err := mockHelmClient.Enable("one", map[string]any{"key": "value"})
	g.Expect(err).To(BeNil())
}
