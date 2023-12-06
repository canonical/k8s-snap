package component

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/componentmanager"
)

const (
	kubeSystemNamespace = "kube-system"
)

// ChartInfo represents the information needed to install a Helm chart.
type ChartInfo struct {
	ReleaseName string
	ChartPath   string
}

// TODO: Include missing components
// componentMap maps a component to its corresponding ChartInfo.
var componentMap = map[string]ChartInfo{
	"cni": {ReleaseName: "ck-cni", ChartPath: path.Join(os.Getenv("SNAP"), "cilium-1.14.1.tgz")},
	"dns": {ReleaseName: "ck-dns", ChartPath: path.Join(os.Getenv("SNAP"), "coredns-1.28.2.tgz")},
}

// DisableComponent uninstalls the specified component using Helm.
func DisableComponent(component string) error {
	chartInfo, ok := componentMap[component]
	if !ok {
		return fmt.Errorf("invalid component: %s.", component)
	}

	client, err := initializeComponentManager()
	if err != nil {
		return err
	}

	enabled, err := client.IsComponentEnabled(chartInfo.ReleaseName, kubeSystemNamespace)
	if err != nil {
		return err
	}

	if !enabled {
		return fmt.Errorf("component %s is not installed or has already been disabled.", component)
	}

	err = client.UninstallComponent(chartInfo.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %w", component, err)
	}
	return nil
}

// EnableComponent installs the specified component using Helm.
func EnableComponent(component string, values map[string]any) error {
	chartInfo, ok := componentMap[component]
	if !ok {
		return fmt.Errorf("invalid component: %s.", component)
	}

	client, err := initializeComponentManager()
	if err != nil {
		return err
	}

	enabled, err := client.IsComponentEnabled(chartInfo.ReleaseName, kubeSystemNamespace)
	if err != nil {
		return err
	}

	if enabled {
		return fmt.Errorf("component %s has already been enabled.", component)
	}

	err = client.InstallComponent(chartInfo.ChartPath, chartInfo.ReleaseName, kubeSystemNamespace, values)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w", component, err)
	}

	return nil
}

// ListEnabledComponents lists all components that are currently enabled.
func ListEnabledComponents() ([]*componentmanager.ComponentSpec, error) {
	client, err := initializeComponentManager()
	if err != nil {
		return nil, fmt.Errorf("error initializing Helm client: %w", err)
	}

	components := make(map[string]bool)
	for name := range componentMap {
		components[name] = true
	}

	enabledComponents, err := client.ListInstalledComponents(components)
	if err != nil {
		return nil, err
	}

	return enabledComponents, nil
}

// initializeComponentManager creates and returns a new Helm client.
func initializeComponentManager() (componentmanager.ComponentManager, error) {
	client, err := componentmanager.NewHelm()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise Component Manager: %w", err)
	}
	return client, nil
}
