package helm

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

// Helm defines the interface for Helm operations
type Helm interface {
	ConfigureRelease(releaseName, namespace string, reuseValue bool, newValues map[string]any) error
	InstallChart(chartPath, releaseName, namespace string, values map[string]any) error
	ListReleases() ([]*release.Release, error)
	UninstallRelease(releaseName string) error
	UpgradeRelease(releaseName, chartPath, namespace string) error
}

// helmClient implements the Helm interface
type helmClient struct {
	settings *cli.EnvSettings
	config   *action.Configuration
}

func logAdapter(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}

// NewHelm creates a new Helm client with the provided settings and configuration
func NewHelm() (Helm, error) {
	settings := cli.New()
	settings.KubeConfig = "/etc/kubernetes/admin.conf"
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
func (h *helmClient) InstallChart(chartPath, releaseName, namespace string, values map[string]interface{}) error {
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

// ListReleases lists all the releases managed by Helm
func (h *helmClient) ListReleases() ([]*release.Release, error) {
	list := action.NewList(h.config)
	releases, err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm releases: %w", err)
	}
	return releases, nil
}

// UninstallRelease uninstalls a specified Helm release
func (h *helmClient) UninstallRelease(releaseName string) error {
	uninstall := action.NewUninstall(h.config)
	if _, err := uninstall.Run(releaseName); err != nil {
		return fmt.Errorf("failed to uninstall release '%s': %w", releaseName, err)
	}
	return nil
}

// UpgradeRelease upgrades a specified Helm release, reusing its values.
func (h *helmClient) UpgradeRelease(releaseName, chartPath, namespace string) error {
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
func (h *helmClient) ConfigureRelease(releaseName, namespace string, reuseValue bool, newValues map[string]any) error {
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
