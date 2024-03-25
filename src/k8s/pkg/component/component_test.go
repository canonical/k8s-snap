package component

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	componentmock "github.com/canonical/k8s/pkg/component/mock"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
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
		Log:            logAdapter,
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

func withName(name string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Name = name
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

func createComponentMap() map[string]types.Component {
	return map[string]types.Component{
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
	}
}

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

func mustAddChartToTestDir(t *testing.T, path string, chart *chart.Chart) string {
	// Create a chart and add it to the test directory as a gzip archive
	k8sComponentsDir := filepath.Join(path, "k8s", "components", "charts")
	chartPath, err := chartutil.Save(chart, k8sComponentsDir)
	if err != nil {
		t.Fatal(err)
	}
	return chartPath
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

	//Create a mock ComponentManager with the mock HelmClient
	mockHelmCLient, err := NewHelmClient(snap, mockClient)
	if err != nil {
		t.Fatal(err)
	}

	return mockHelmCLient, tempDir, mockActionConfig
}

func TestListEmptyComponents(t *testing.T) {
	g := NewWithT(t)
	// Create a mock ComponentManager with no components
	mockHelmClient, _, _ := mustCreateNewHelmClient(t, nil)

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(components).To(HaveLen(0))
}

func TestListComponentsWithReleases(t *testing.T) {
	g := NewWithT(t)

	// Create a mock ComponentManager with the mock HelmClient
	// This mock uses components.yaml for the snap mock components
	mockHelmClient, _, mockActionConfig := mustCreateNewHelmClient(t, createComponentMap())

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

func TestComponentsInitialState(t *testing.T) {
	g := NewWithT(t)

	components := createComponentMap()

	mockHelmClient, _, _ := mustCreateNewHelmClient(t, components)

	for _, component := range components {
		t.Run(component.ReleaseName, func(t *testing.T) {
			g.Expect(mockHelmClient.isComponentEnabled(component.ReleaseName, component.Namespace)).To(BeFalse(), "Expected all components to be initially disabled")
		})
	}
}

func TestEnableMultipleComponents(t *testing.T) {
	g := NewWithT(t)

	components := createComponentMap()

	mockHelmClient, tempDir, _ := mustCreateNewHelmClient(t, components)

	for name, component := range mockHelmClient.components {
		chart := buildChart(withName(component.ReleaseName))
		chartPath := mustAddChartToTestDir(t, tempDir, chart)
		component.ManifestPath = chartPath
		mockHelmClient.components[name] = component

		err := mockHelmClient.Enable(name, map[string]interface{}{})
		g.Expect(err).ShouldNot(HaveOccurred())
	}

	// Reenable should not error
	t.Run("Reenable", func(t *testing.T) {
		for name := range mockHelmClient.components {
			g.Expect(mockHelmClient.Enable(name, map[string]any{})).Should(Succeed())
		}
	})

	for _, component := range components {
		t.Run(component.ReleaseName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(mockHelmClient.isComponentEnabled(component.ReleaseName, component.Namespace)).To(BeTrue(), "Expected all components to enabled")
		})
	}

	// Component does not exist at all
	t.Run("EnableNonExistent", func(t *testing.T) {
		g.Expect(mockHelmClient.Enable("non-existent", map[string]any{})).ShouldNot(Succeed())
	})
}

func TestDisableComponent(t *testing.T) {
	g := NewWithT(t)

	components := createComponentMap()

	mockHelmClient, tempDir, _ := mustCreateNewHelmClient(t, components)

	for name, component := range mockHelmClient.components {
		chart := buildChart(withName(component.ReleaseName))
		chartPath := mustAddChartToTestDir(t, tempDir, chart)
		component.ManifestPath = chartPath
		mockHelmClient.components[name] = component

		err := mockHelmClient.Enable(name, map[string]interface{}{})
		g.Expect(err).ShouldNot(HaveOccurred())
	}

	// Redisable should not error
	t.Run("Redisable", func(t *testing.T) {
		for name := range mockHelmClient.components {
			g.Expect(mockHelmClient.Disable(name)).Should(Succeed())
		}
	})

	for _, component := range components {
		t.Run(component.ReleaseName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(mockHelmClient.isComponentEnabled(component.ReleaseName, component.ReleaseName)).To(BeFalse(), "Expected all components to be disabled")
		})
	}

	// Component does not exist at all
	t.Run("DisableNonExistent", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(mockHelmClient.Disable("non-existent")).ShouldNot(Succeed())
	})
}

func TestRefresh(t *testing.T) {
	g := NewWithT(t)
	components := map[string]types.Component{
		"whiskas-1": {
			ReleaseName:  "whiskas-1",
			Namespace:    "default",
			ManifestPath: "chunky-tuna-1.14.1.tgz",
		},
	}

	mockHelmClient, tempDir, _ := mustCreateNewHelmClient(t, components)

	for name, component := range mockHelmClient.components {
		chart := buildChart(withName(component.ReleaseName))
		chartPath := mustAddChartToTestDir(t, tempDir, chart)
		component.ManifestPath = chartPath
		mockHelmClient.components[name] = component
		g.Expect(mockHelmClient.Enable(name, map[string]interface{}{})).To(Succeed())
	}

	t.Run("RefreshExisting", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(mockHelmClient.Refresh("whiskas-1", map[string]interface{}{})).To(gomega.Succeed())
	})

	t.Run("RefreshNonExisting", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(mockHelmClient.Refresh("non-existent", map[string]interface{}{})).ToNot(Succeed())
	})
}
