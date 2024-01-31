package component

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmConfigInitializer interface {
	// Initializes a fresh Helm Configuration
	InitializeHelmClientConfig() (*action.Configuration, error)
}

type HelmClientIntitializer struct{}

func (r *HelmClientIntitializer) InitializeHelmClientConfig() (*action.Configuration, error) {
	settings := cli.New()
	settings.KubeConfig = "/etc/kubernetes/admin.conf"
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		logAdapter,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize action config: %w", err)
	}
	return actionConfig, nil
}

// ComponentManager defines an interface for managing k8s components.
type ComponentManager interface {
	// Enable enables a k8s component, optionally specifying custom configuration options.
	Enable(name string, values map[string]any) error
	// List returns a list of enabled components.
	List() ([]Component, error)
	// Disable disables a component from the cluster.
	Disable(name string) error
	// Refresh updates a k8s component.
	Refresh(name string) error
}

// componentDefinition defines each component metadata.
type componentDefinition struct {
	ParentComponent string `mapstructure:"parent"`
	ReleaseName     string `mapstructure:"release"`
	Chart           string `mapstructure:"chart"`
	Namespace       string `mapstructure:"namespace"`
}

// helmClient implements the ComponentManager interface
type helmClient struct {
	config      map[string]componentDefinition
	snap        snap.Snap
	initializer HelmConfigInitializer
}

// Component defines the name and status of a k8s Component.
type Component struct {
	Name   string
	Status bool
}

func logAdapter(format string, v ...any) {
	logrus.Debugf(format, v...)
}

// NewManager creates a new Component manager instance.
func NewManager(snap snap.Snap, initializer HelmConfigInitializer) (*helmClient, error) {
	if initializer == nil {
		initializer = &HelmClientIntitializer{}
	}

	viper.SetConfigName("components")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(snap.Path("k8s/components"))
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := make(map[string]componentDefinition)
	err = viper.Unmarshal(&config)

	return &helmClient{
		config:      config,
		snap:        snap,
		initializer: initializer,
	}, nil
}

// Enable enables a specified component.
func (h *helmClient) Enable(name string, values map[string]any) error {
	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	actionConfig, err := h.initializer.InitializeHelmClientConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	install := action.NewInstall(actionConfig)
	install.ReleaseName = component.ReleaseName
	install.Namespace = component.Namespace

	isEnabled, err := h.isComponentEnabled(component.ReleaseName, component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get components status: %w", err)
	}

	if isEnabled {
		return nil
	}
	chart, err := loader.Load(h.snap.Path("k8s/components/charts", component.Chart))
	if err != nil {
		return fmt.Errorf("failed to load component manifest: %w", err)
	}
	_, err = install.Run(chart, values)
	if err != nil {
		return fmt.Errorf("failed to enable component '%s': %w", name, err)
	}

	return nil
}

// isComponentEnabled checks if a component is enabled.
func (h *helmClient) isComponentEnabled(name, namespace string) (bool, error) {
	actionConfig, err := h.initializer.InitializeHelmClientConfig()
	if err != nil {
		return false, fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	list := action.NewList(actionConfig)
	releases, err := list.Run()
	if err != nil {
		return false, err
	}

	for _, release := range releases {
		if release.Name == name && release.Namespace == namespace {
			return true, nil
		}
	}

	return false, nil
}

// List lists the status of each k8s component.
func (h *helmClient) List() ([]Component, error) {
	actionConfig, err := h.initializer.InitializeHelmClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	list := action.NewList(actionConfig)
	releases, err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	allComponents := make([]Component, len(h.config))
	componentsMap := make(map[string]int)
	for name, component := range h.config {
		index := len(componentsMap)
		allComponents[index] = Component{Name: name}
		componentsMap[component.ReleaseName] = index
	}

	for _, release := range releases {
		index, ok := componentsMap[release.Name]
		if ok {
			allComponents[index].Status = true
		}
	}

	return allComponents, nil
}

// Disable disables a specified component.
func (h *helmClient) Disable(name string) error {
	actionConfig, err := h.initializer.InitializeHelmClientConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	uninstall := action.NewUninstall(actionConfig)
	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	isEnabled, err := h.isComponentEnabled(component.ReleaseName, component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get components status: %w", err)
	}

	if !isEnabled {
		return nil
	}
	_, err = uninstall.Run(component.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall component '%s': %w", name, err)
	}

	return nil
}

// Refresh refreshes a specified component.
func (h *helmClient) Refresh(name string, values map[string]any) error {
	actionConfig, err := h.initializer.InitializeHelmClientConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = component.Namespace
	upgrade.ReuseValues = true

	chart, err := loader.Load(h.snap.Path("k8s/components/charts", component.Chart))
	if err != nil {
		return fmt.Errorf("failed to load component manifest: %w", err)
	}

	_, err = upgrade.Run(component.ReleaseName, chart, values)
	if err != nil {
		return fmt.Errorf("failed to upgrade component '%s': %w", name, err)
	}
	return nil
}
