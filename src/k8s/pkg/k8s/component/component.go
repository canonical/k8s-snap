package component

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/helm"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
)

const (
	errInitHelm           = "failed to initialize Helm: %w"
	errInstallComponent   = "failed to install %s: %w"
	errUninstallComponent = "failed to uninstall %s: %w"
	snapDir               = "/snap/k8s/current/"
	kubeSystemNamespace   = "kube-system"
)

// ChartInfo represents the information needed to install a Helm chart.
type ChartInfo struct {
	ReleaseName string
	ChartPath   string
}

// TODO: Include missing components
// componentMap maps a component to its corresponding ChartInfo.
var componentMap = map[string]ChartInfo{
	"cni": {ReleaseName: "ck-cni", ChartPath: snapDir + "cilium"},
	"dns": {ReleaseName: "ck-dns", ChartPath: snapDir + "coredns"},
}

// DisableComponent uninstalls the specified component using Helm.
func DisableComponent(component string) error {
	chartInfo, ok := componentMap[component]
	if !ok {
		return fmt.Errorf("invalid component: %s.", component)
	}

	client, err := initializeHelmClient()
	if err != nil {
		return err
	}

	releases, err := client.ListReleases()
	if err != nil {
		return err
	}

	if !checkRelease(chartInfo.ReleaseName, releases) {
		return fmt.Errorf("component %s is not installed or has already been disabled.", component)
	}

	err = client.UninstallRelease(chartInfo.ReleaseName)
	if err != nil {
		return fmt.Errorf(errUninstallComponent, component, err)
	}
	return nil
}

// EnableComponent installs the specified component using Helm.
func EnableComponent(component string, values map[string]interface{}, configFile string) error {
	chartInfo, ok := componentMap[component]
	if !ok {
		return fmt.Errorf("invalid component: %s.", component)
	}

	client, err := initializeHelmClient()
	if err != nil {
		return err
	}

	var chartValues map[string]interface{}
	if configFile != "" {
		chartValues, err = readYAMLFile(configFile)
		if err != nil {
			return err
		}
	} else {
		chartValues = values
	}

	releases, err := client.ListReleases()
	if err != nil {
		return err
	}

	if checkRelease(chartInfo.ReleaseName, releases) {
		return fmt.Errorf("component %s has already been enabled.", component)
	}

	err = client.InstallChart(chartInfo.ChartPath, chartInfo.ReleaseName, kubeSystemNamespace, chartValues)
	if err != nil {
		return fmt.Errorf(errInstallComponent, component, err)
	}

	return nil
}

// ListEnabledComponents lists all components that are currently enabled.
func ListEnabledComponents() ([]string, error) {
	client, err := initializeHelmClient()
	if err != nil {
		return nil, fmt.Errorf("error initializing Helm client: %w", err)
	}

	releases, err := client.ListReleases()
	if err != nil {
		return nil, fmt.Errorf("error listing Helm releases: %w", err)
	}

	releaseSet := make(map[string]bool)
	for _, r := range releases {
		releaseSet[r.Name] = true
	}

	var enabledComponents []string
	for component, chartInfo := range componentMap {
		if releaseSet[chartInfo.ReleaseName] {
			enabledComponents = append(enabledComponents, component)
		}
	}

	return enabledComponents, nil
}

// checkRelease checks if a Helm release with the given releaseName is present in the list of releases.
func checkRelease(releaseName string, releases []*release.Release) bool {
	for _, rel := range releases {
		if rel.Name == releaseName {
			return true
		}
	}
	return false
}

// initializeHelmClient creates and returns a new Helm client.
func initializeHelmClient() (helm.Helm, error) {
	client, err := helm.NewHelm()
	if err != nil {
		return nil, fmt.Errorf(errInitHelm, err)
	}
	return client, nil
}

// readYAMLFile read a YAML file from the given path an returns its content.
func readYAMLFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	sanitisedMap := sanitiseMap(content)

	return sanitisedMap, nil
}

// sanitiseMap converts a map with interface{} keys to a map with string keys.
// This is useful for preparing data for use with the Helm client, which requires
// map keys to be strings. Nested maps are also recursively processed to ensure
// all keys are converted to strings.
func sanitiseMap(m map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		switch t := value.(type) {
		case map[interface{}]interface{}:
			result[fmt.Sprint(key)] = sanitiseMap(t)
		default:
			result[fmt.Sprint(key)] = value
		}
	}
	return result
}
