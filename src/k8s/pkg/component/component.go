package component

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

// ComponentManager defines an interface for managing k8s components.
type ComponentManager interface {
	// Enable enables a k8s component.
	Enable(name string) error
	// List returns a list of enabled components.
	List() ([]Component, error)
	// Disable disables a component from the cluster.
	Disable(name string) error
	// Refresh updates a k8s component.
	Refresh(name string) error
}

// componentDefinition defines each component metadata.
type componentDefinition struct {
	ReleaseName string `mapstructure:"release"`
	Chart       string `mapstructure:"chart"`
	Namespace   string `mapstructure:"namespace"`
}

// helmClient implements the ComponentManager interface
type helmClient struct {
	config       map[string]componentDefinition
	settings     *cli.EnvSettings
	actionConfig *action.Configuration
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
func NewManager() (*helmClient, error) {
	viper.SetConfigName("components")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(utils.Path("k8s/components"))
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := make(map[string]componentDefinition)
	err = viper.Unmarshal(&config)

	settings := cli.New()
	settings.KubeConfig = "/etc/kubernetes/admin.conf"
	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		logAdapter,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize component manager configuration: %w", err)
	}

	return &helmClient{
		config:       config,
		settings:     settings,
		actionConfig: actionConfig,
	}, nil
}

// Enable enables a specified component.
func (h *helmClient) Enable(name string) error {
	install := action.NewInstall(h.actionConfig)
	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}
	install.ReleaseName = component.ReleaseName
	install.Namespace = component.Namespace

	isEnabled, err := h.isComponentEnabled(name, component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get components status: %w", err)
	}

	if isEnabled {
		return nil
	}

	chart, err := loader.Load(utils.Path("k8s/components/charts", component.Chart))
	if err != nil {
		return fmt.Errorf("failed to load component manifest: %w", err)
	}

	_, err = install.Run(chart, nil)
	if err != nil {
		return fmt.Errorf("failed to enable component '%s': %w", name, err)
	}
	return nil
}

// isComponentEnabled checks if a component is enabled.
func (h *helmClient) isComponentEnabled(name, namespace string) (bool, error) {
	list := action.NewList(h.actionConfig)
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
	list := action.NewList(h.actionConfig)
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
	uninstall := action.NewUninstall(h.actionConfig)
	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	isEnabled, err := h.isComponentEnabled(name, component.Namespace)
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
func (h *helmClient) Refresh(name string) error {
	component, ok := h.config[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	upgrade := action.NewUpgrade(h.actionConfig)
	upgrade.Namespace = component.Namespace
	upgrade.ReuseValues = true

	chart, err := loader.Load(utils.Path(component.Chart))
	if err != nil {
		return fmt.Errorf("failed to load component manifest: %w", err)
	}

	_, err = upgrade.Run(component.ReleaseName, chart, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade component '%s': %w", name, err)
	}
	return nil
}
