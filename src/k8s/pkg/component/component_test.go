package component

import (
	"flag"
	"io"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
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

// "Mock" Initializer
type MockHelmClientInitializer struct {
	actionConfig *action.Configuration
}

func (r *MockHelmClientInitializer) InitializeHelmClientConfig() (*action.Configuration, error) {
	return r.actionConfig, nil
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

	// Call the List function with the mock HelmClient
	components, err := mockHelmClient.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(components))
	}
}
