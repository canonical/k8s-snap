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

var manifestWithHook = `kind: ConfigMap
metadata:
  name: test-cm
  annotations:
    "helm.sh/hook": post-install,pre-delete,post-upgrade
data:
  name: value`

var manifestWithTestHook = `kind: Pod
  metadata:
	name: finding-nemo,
	annotations:
	  "helm.sh/hook": test
  spec:
	containers:
	- name: nemo-test
	  image: fake-image
	  cmd: fake-command
  `

var rbacManifests = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: schedule-agents
rules:
- apiGroups: [""]
  resources: ["pods", "pods/exec", "pods/log"]
  verbs: ["*"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: schedule-agents
  namespace: {{ default .Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: schedule-agents
subjects:
- kind: ServiceAccount
  name: schedule-agents
  namespace: {{ .Release.Namespace }}
`

type chartOptions struct {
	*chart.Chart
}

type chartOption func(*chartOptions)

func buildChart(opts ...chartOption) *chart.Chart {
	c := &chartOptions{
		Chart: &chart.Chart{
			// TODO: This should be more complete.
			Metadata: &chart.Metadata{
				APIVersion: "v1",
				Name:       "hello",
				Version:    "0.1.0",
			},
			// This adds a basic template and hooks.
			Templates: []*chart.File{
				{Name: "templates/hello", Data: []byte("hello: world")},
				{Name: "templates/hooks", Data: []byte(manifestWithHook)},
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

func withSampleValues() chartOption {
	values := map[string]interface{}{
		"someKey": "someValue",
		"nestedKey": map[string]interface{}{
			"simpleKey": "simpleValue",
			"anotherNestedKey": map[string]interface{}{
				"yetAnotherNestedKey": map[string]interface{}{
					"youReadyForAnotherNestedKey": "No",
				},
			},
		},
	}
	return func(opts *chartOptions) {
		opts.Values = values
	}
}

func withValues(values map[string]interface{}) chartOption {
	return func(opts *chartOptions) {
		opts.Values = values
	}
}

func withNotes(notes string) chartOption {
	return func(opts *chartOptions) {
		opts.Templates = append(opts.Templates, &chart.File{
			Name: "templates/NOTES.txt",
			Data: []byte(notes),
		})
	}
}

func withDependency(dependencyOpts ...chartOption) chartOption {
	return func(opts *chartOptions) {
		opts.AddDependency(buildChart(dependencyOpts...))
	}
}

func withMetadataDependency(dependency chart.Dependency) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Dependencies = append(opts.Metadata.Dependencies, &dependency)
	}
}

func withSampleTemplates() chartOption {
	return func(opts *chartOptions) {
		sampleTemplates := []*chart.File{
			// This adds basic templates and partials.
			{Name: "templates/goodbye", Data: []byte("goodbye: world")},
			{Name: "templates/empty", Data: []byte("")},
			{Name: "templates/with-partials", Data: []byte(`hello: {{ template "_planet" . }}`)},
			{Name: "templates/partials/_planet", Data: []byte(`{{define "_planet"}}Earth{{end}}`)},
		}
		opts.Templates = append(opts.Templates, sampleTemplates...)
	}
}

func withSampleIncludingIncorrectTemplates() chartOption {
	return func(opts *chartOptions) {
		sampleTemplates := []*chart.File{
			// This adds basic templates and partials.
			{Name: "templates/goodbye", Data: []byte("goodbye: world")},
			{Name: "templates/empty", Data: []byte("")},
			{Name: "templates/incorrect", Data: []byte("{{ .Values.bad.doh }}")},
			{Name: "templates/with-partials", Data: []byte(`hello: {{ template "_planet" . }}`)},
			{Name: "templates/partials/_planet", Data: []byte(`{{define "_planet"}}Earth{{end}}`)},
		}
		opts.Templates = append(opts.Templates, sampleTemplates...)
	}
}

func withMultipleManifestTemplate() chartOption {
	return func(opts *chartOptions) {
		sampleTemplates := []*chart.File{
			{Name: "templates/rbac", Data: []byte(rbacManifests)},
		}
		opts.Templates = append(opts.Templates, sampleTemplates...)
	}
}

func withKube(version string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.KubeVersion = version
	}
}

// releaseStub creates a release stub, complete with the chartStub as its chart.
func releaseStub() *release.Release {
	return namedReleaseStub("angry-panda", release.StatusDeployed)
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
		Chart:   buildChart(withSampleTemplates()),
		Config:  map[string]interface{}{"name": "value"},
		Version: 1,
		Hooks: []*release.Hook{
			{
				Name:     "test-cm",
				Kind:     "ConfigMap",
				Path:     "test-cm",
				Manifest: manifestWithHook,
				Events: []release.HookEvent{
					release.HookPostInstall,
					release.HookPreDelete,
				},
			},
			{
				Name:     "finding-nemo",
				Kind:     "Pod",
				Path:     "finding-nemo",
				Manifest: manifestWithTestHook,
				Events: []release.HookEvent{
					release.HookTest,
				},
			},
		},
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
	one := releaseStub()
	one.Name = "one"
	one.Namespace = "default"
	one.Version = 1
	two := releaseStub()
	two.Name = "two"
	two.Namespace = "default"
	two.Version = 2
	three := releaseStub()
	three.Name = "three"
	three.Namespace = "default"
	three.Version = 3

	for _, rel := range []*release.Release{one, two, three} {
		if err := store.Create(rel); err != nil {
			t.Fatal(err)
		}
	}

	all, err := store.ListReleases()
	assert.NoError(t, err)
	assert.Len(t, all, 3, "sanity test: three items added")
}

var components = `
network:
  release: "ck-network"
  chart: "cilium-1.14.1.tgz"
  namespace: "kube-system"
dns:
  release: "ck-dns"
  chart: "coredns-1.29.0.tgz"
  namespace: "kube-system"
`

func createTemporaryTestDirectory(t *testing.T) string {
	// Create a temporary test directory
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

func TestNewManager(t *testing.T) {
	// Create a mock actionConfig for testing
	mockActionConfig := actionConfigFixture(t)
	// Create a mock HelmClient with the desired behavior for testing
	mockClient := &MockHelmClientInitializer{actionConfig: mockActionConfig}

	// create test directory
	tempDir := createTemporaryTestDirectory(t)
	defer os.RemoveAll(tempDir)

	// Create a file and add some configs
	addConfigToTestDir(t, filepath.Join(tempDir, "k8s", "components", "components.yaml"), components)

	// Create mock snap
	snap := &mock.Snap{
		PathPrefix: tempDir, //make a test dir?
	}
	//Create a mock ComponentManager with the mock HelmClient
	mockComponentManager, err := NewManager(snap, mockClient)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, mockComponentManager)
	assert.IsType(t, &helmClient{}, mockComponentManager)
}

func TestListEmpty(t *testing.T) {
	// Create a mock actionConfig for testing
	mockActionConfig := actionConfigFixture(t)
	// Create a mock HelmClient with the desired behavior for testing
	mockClient := &MockHelmClientInitializer{actionConfig: mockActionConfig}

	//Create a mock ComponentManager with the mock HelmClient
	mockHelmClient := &helmClient{
		initializer: mockClient,
	}

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
	// Create a mock actionConfig for testing
	mockActionConfig := actionConfigFixture(t)
	// Create a mock HelmClient with the desired behavior for testing
	mockClient := &MockHelmClientInitializer{actionConfig: mockActionConfig}

	//Create a mock ComponentManager with the mock HelmClient
	mockHelmClient := &helmClient{
		initializer: mockClient,
	}

	// Create releases in the mock actionConfig
	makeMeSomeReleases(mockActionConfig.Releases, t)

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, components)
}
