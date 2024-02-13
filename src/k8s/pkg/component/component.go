package component

import (
	"fmt"
	"os"
	"sort"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// defaultHelmConfigProvider implements the HelmConfigInitializer interface
type defaultHelmConfigProvider struct {
	restClientGetter func(namespace string) genericclioptions.RESTClientGetter
}

// helmClient implements the ComponentManager interface
type helmClient struct {
	components  map[string]types.Component
	initializer HelmConfigProvider
}

// Component defines the name and status of a k8s Component.
type Component struct {
	Name   string
	Status bool
}

// InitializeHelmClientConfig initializes a Helm Configuration, ensures the use of a fresh configuration
func (r *defaultHelmConfigProvider) New(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(r.restClientGetter(namespace), namespace, os.Getenv("HELM_DRIVER"), logAdapter); err != nil {
		return nil, fmt.Errorf("failed to initialize action config: %w", err)
	}
	return actionConfig, nil
}

func logAdapter(format string, v ...any) {
	logrus.Debugf(format, v...)
}

// NewHelmClient creates a new Component manager instance.
func NewHelmClient(snap snap.Snap, initializer HelmConfigProvider) (*helmClient, error) {
	if initializer == nil {
		initializer = &defaultHelmConfigProvider{restClientGetter: snap.KubernetesRESTClientGetter}
	}

	return &helmClient{
		components:  snap.Components(),
		initializer: initializer,
	}, nil
}

// Enable enables a specified component.
func (h *helmClient) Enable(name string, values map[string]any) error {
	component, ok := h.components[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	actionConfig, err := h.initializer.New(component.Namespace)
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

	chart, err := loader.Load(component.ManifestPath)
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
	actionConfig, err := h.initializer.New(namespace)
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
	actionConfig, err := h.initializer.New("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	list := action.NewList(actionConfig)
	releases, err := list.Run()

	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	allComponents := make([]Component, 0, len(h.components))
	componentsMap := make(map[string]int)

	// Loop through components and populate allComponents and componentsMap
	for name, component := range h.components {
		index := len(componentsMap)

		allComponents = append(allComponents, Component{Name: name})
		componentsMap[component.ReleaseName] = index
	}

	// Loop through releases and update statuses in allComponents
	for _, release := range releases {
		if index, ok := componentsMap[release.Name]; ok {
			allComponents[index].Status = true
		}
	}

	sort.Slice(allComponents, func(i, j int) bool {
		return allComponents[i].Name < allComponents[j].Name
	})

	return allComponents, nil
}

// Disable disables a specified component.
func (h *helmClient) Disable(name string) error {
	component, ok := h.components[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	actionConfig, err := h.initializer.New(component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	isEnabled, err := h.isComponentEnabled(component.ReleaseName, component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get components status: %w", err)
	}

	if !isEnabled {
		return nil
	}

	uninstall := action.NewUninstall(actionConfig)
	_, err = uninstall.Run(component.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall component '%s': %w", name, err)
	}

	return nil
}

// Refresh refreshes a specified component.
func (h *helmClient) Refresh(name string, values map[string]any) error {
	component, ok := h.components[name]
	if !ok {
		return fmt.Errorf("invalid component %s", name)
	}

	actionConfig, err := h.initializer.New(component.Namespace)
	if err != nil {
		return fmt.Errorf("failed to initialize Helm client configuration: %w", err)
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = component.Namespace
	upgrade.ReuseValues = true

	chart, err := loader.Load(component.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to load component manifest: %w", err)
	}

	_, err = upgrade.Run(component.ReleaseName, chart, values)
	if err != nil {
		return fmt.Errorf("failed to upgrade component '%s': %w", name, err)
	}
	return nil
}
