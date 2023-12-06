package componentmanager

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

// ComponentManager defines an interface for managing k8s components.
type ComponentManager interface {
	// ConfigureComponent updates configurations of an existing component.
	ConfigureComponent(releaseName, namespace string, reuseValue bool, newValues map[string]any) error
	// InstallComponent installs a component in a given namespace.
	InstallComponent(chartPath, releaseName, namespace string, values map[string]any) error
	// IsComponentEnabled checks if a component is currently active in the cluster.
	IsComponentEnabled(releaseName, namespace string) (bool, error)
	// ListInstalledComponents returns a list of enabled components based on the available components.
	ListInstalledComponents(components map[string]bool) ([]*ComponentSpec, error)
	// UninstallComponent removes a component from the cluster.
	UninstallComponent(releaseName string) error
	// UpgradeComponent updates a component with a new chart.
	UpgradeComponent(releaseName, chartPath, namespace string) error
}

// helmClient implements the ComponentManager interface
type helmClient struct {
	settings *cli.EnvSettings
	config   *action.Configuration
}

// ComponentSpec defines the specifications of a component, including its configuration and metadata.
type ComponentSpec struct {
	Name        string
	Chart       string
	Namespace   string
	Values      map[string]any
	ReuseValues bool
}

func logAdapter(format string, v ...any) {
	logrus.Debugf(format, v...)
}

// NewHelm creates a new Helm client with the provided settings and configuration
func NewHelm() (*helmClient, error) {
	settings := cli.New()
	settings.KubeConfig = "/var/snap/k8s/common/etc/kubernetes/admin.conf"
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), logAdapter); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm configuration: %w", err)
	}

	return &helmClient{
		settings: settings,
		config:   actionConfig,
	}, nil
}

// InstallChart installs a Helm chart into the specified namespace with the given values
func (h *helmClient) InstallComponent(chartPath, releaseName, namespace string, values map[string]any) error {
	install := action.NewInstall(h.config)
	install.ReleaseName = releaseName
	install.Namespace = namespace

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart '%s': %w", chartPath, err)
	}

	if _, err := install.Run(chart, values); err != nil {
		return fmt.Errorf("failed to install chart '%s': %w", chartPath, err)
	}
	return nil
}

// IsComponentEnabled checks if a component with the given release name is enabled.
func (h *helmClient) IsComponentEnabled(releaseName, namespace string) (bool, error) {
	list := action.NewList(h.config)
	releases, err := list.Run()
	if err != nil {
		return false, fmt.Errorf("failed to list Helm releases: %w", err)
	}

	for _, release := range releases {
		if release.Name == releaseName && release.Namespace == namespace {
			return true, nil
		}
	}

	return false, nil
}

// ListReleases lists all the releases managed by Helm
func (h *helmClient) ListInstalledComponents(components map[string]bool) ([]*ComponentSpec, error) {
	list := action.NewList(h.config)
	releases, err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm releases: %w", err)
	}

	var enabledComponents []*ComponentSpec
	for _, release := range releases {
		if components[release.Name] {
			component := &ComponentSpec{
				Name:      release.Name,
				Chart:     release.Chart.ChartPath(),
				Namespace: release.Namespace,
				Values:    release.Config,
			}
			enabledComponents = append(enabledComponents, component)
		}
	}

	return enabledComponents, nil
}

// UninstallRelease uninstalls a specified Helm release
func (h *helmClient) UninstallComponent(releaseName string) error {
	uninstall := action.NewUninstall(h.config)
	if _, err := uninstall.Run(releaseName); err != nil {
		return fmt.Errorf("failed to uninstall release '%s': %w", releaseName, err)
	}
	return nil
}

// UpgradeRelease upgrades a specified Helm release, reusing its values.
func (h *helmClient) UpgradeComponent(releaseName, chartPath, namespace string) error {
	upgrade := action.NewUpgrade(h.config)
	upgrade.Namespace = namespace
	upgrade.ReuseValues = true

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart '%s': %w", chartPath, err)
	}

	_, err = upgrade.Run(releaseName, chart, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade release '%s': %w", releaseName, err)
	}
	return nil
}

// ConfigureRelease configures a specified Helm release with the given values.
func (h *helmClient) ConfigureComponent(releaseName, namespace string, reuseValue bool, newValues map[string]any) error {
	upgrade := action.NewUpgrade(h.config)
	upgrade.Namespace = namespace

	get := action.NewGet(h.config)
	release, err := get.Run(releaseName)
	if err != nil {
		return fmt.Errorf("unable to get release %s: %w", releaseName, err)
	}

	upgrade.ReuseValues = reuseValue
	_, err = upgrade.Run(releaseName, release.Chart, newValues)
	if err != nil {
		return fmt.Errorf("failed to reconfigure release %s: %w", releaseName, err)
	}

	return nil
}
