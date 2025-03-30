package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const FEATURE_VERSION_LABEL = "k8sd.io/%s-version"

// client implements Client using Helm.
type client struct {
	restClientGetter func(string) genericclioptions.RESTClientGetter
	manifestsBaseDir string
}

// ensure *client implements Client.
var _ Client = &client{}

// NewClient creates a new client.
func NewClient(manifestsBaseDir string, restClientGetter func(string) genericclioptions.RESTClientGetter) *client {
	return &client{
		restClientGetter: restClientGetter,
		manifestsBaseDir: manifestsBaseDir,
	}
}

func (h *client) newActionConfiguration(ctx context.Context, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	log := log.FromContext(ctx).WithName("helm")
	if err := actionConfig.Init(h.restClientGetter(namespace), namespace, "", func(format string, v ...interface{}) {
		log.Info(fmt.Sprintf(format, v...))
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}
	return actionConfig, nil
}

// Apply implements the Client interface.
func (h *client) Apply(ctx context.Context, feature types.FeatureName, version string, c InstallableChart, desired State, values map[string]any) (bool, error) {
	cfg, err := h.newActionConfiguration(ctx, c.Namespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	isInstalled := true
	var oldConfig map[string]any

	// get the latest Helm release with the specified name
	get := action.NewGet(cfg)
	release, err := get.Run(c.Name)
	if err != nil {
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			return false, fmt.Errorf("failed to get status of release %s: %w", c.Name, err)
		}
		isInstalled = false
	} else {
		// keep the existing release configuration, to check if any changes were made.
		oldConfig = release.Config
	}

	var current string
	if release != nil {
		// Get feature version labels from the release
		// This is used to track how the release was installed
		// compare if the release was installed with the same feature version we are trying to use
		// Apply operation is not allowed/blocked if the feature version current snap contains is different
		// An upgrade should be performed before cluster config changes are reconciled
		var ok bool
		current, ok = release.Labels[fmt.Sprintf(FEATURE_VERSION_LABEL, feature)]
		if !ok {
			current = version
		}
	}

	switch {
	case !isInstalled && desired == StateDeleted:
		// no-op
		return false, nil
	case !isInstalled && desired == StateUpgradeOnly:
		// there is no release installed, this is an error
		return false, fmt.Errorf("cannot upgrade %s as it is not installed", c.Name)
	case !isInstalled && desired == StatePresent:
		// there is no release installed, so we must run an install action
		install := action.NewInstall(cfg)
		install.ReleaseName = c.Name
		install.Namespace = c.Namespace
		install.CreateNamespace = true
		// Apply the feature version labels to the release
		install.Labels = map[string]string{
			fmt.Sprintf(FEATURE_VERSION_LABEL, feature): version,
		}

		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		if _, err := install.RunWithContext(ctx, chart, values); err != nil {
			return false, fmt.Errorf("failed to install %s: %w", c.Name, err)
		}
		return true, nil
	case isInstalled && desired != StateDeleted:
		// there is already a release installed, so we must run an upgrade action
		if current != version {
			return false, fmt.Errorf("cannot perform an upgrade operation as this node contains resources for a different version of the feature %s", feature)
		}
		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = c.Namespace
		upgrade.ResetThenReuseValues = true
		// Apply the feature version labels to the release
		upgrade.Labels = map[string]string{
			fmt.Sprintf(FEATURE_VERSION_LABEL, feature): version,
		}

		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		release, err := upgrade.RunWithContext(ctx, c.Name, chart, values)
		if err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", c.Name, err)
		}

		// oldConfig and release.Config are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		return !jsonEqual(oldConfig, release.Config), nil
	case isInstalled && desired == StateDeleted:
		// run an uninstall action
		uninstall := action.NewUninstall(cfg)
		if _, err := uninstall.Run(c.Name); err != nil {
			return false, fmt.Errorf("failed to uninstall %s: %w", c.Name, err)
		}

		return true, nil
	default:
		// this never happens
		return false, nil
	}
}

// Apply implements the Client interface.
func (h *client) ApplyDependent(ctx context.Context, parent FeatureMeta, sub PseudoFeatureMeta, desired State, values map[string]any) (bool, error) {
	c := parent.Chart
	parentFeature := parent.FeatureName
	parentVersion := parent.Version

	subFeature := sub.FeatureName
	subVersion := sub.Version

	cfg, err := h.newActionConfiguration(ctx, c.Namespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	parentIsInstalled := true
	var oldParentConfig map[string]any

	// get the latest Helm release with the specified name
	get := action.NewGet(cfg)
	parentRelease, err := get.Run(c.Name)
	if err != nil {
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			return false, fmt.Errorf("failed to get status of release %s: %w", c.Name, err)
		}
		parentIsInstalled = false
	} else {
		// keep the existing release configuration, to check if any changes were made.
		oldParentConfig = parentRelease.Config
	}

	var parentCurrent string
	if parentRelease != nil {
		// Get feature version labels from the release
		// This is used to track how the release was installed
		// compare if the release was installed with the same feature version we are trying to use
		// Apply operation is not allowed/blocked if the feature version current snap contains is different
		// An upgrade should be performed before cluster config changes are reconciled
		var ok bool
		parentCurrent, ok = parentRelease.Labels[fmt.Sprintf(FEATURE_VERSION_LABEL, parentFeature)]
		if !ok {
			parentCurrent = parentVersion
		}
	}

	if parentCurrent != parentVersion {
		return false, fmt.Errorf("cannot perform an upgrade operation as this node contains resources for a different version of the feature %s", parentFeature)
	}

	var subCurrent string
	if parentRelease != nil {
		// Get feature version labels from the release
		// This is used to track how the release was installed
		// compare if the release was installed with the same feature version we are trying to use
		// Apply operation is not allowed/blocked if the feature version current snap contains is different
		// An upgrade should be performed before cluster config changes are reconciled
		var ok bool
		subCurrent, ok = parentRelease.Labels[fmt.Sprintf(FEATURE_VERSION_LABEL, subFeature)]
		if !ok {
			subCurrent = subVersion
		}
	}

	if subCurrent != subVersion {
		return false, fmt.Errorf("cannot perform an upgrade operation as this node contains resources for a different version of the feature %s", subFeature)
	}

	switch {
	case !parentIsInstalled && desired == StateDeleted:
		// no-op
		return false, nil
	case !parentIsInstalled && desired == StatePresent:
		// the parent feature is not installed, we cannot install the sub feature
		return false, fmt.Errorf("cannot install %s as the parent feature %s is not installed", sub.FeatureName, parentFeature)
	case parentIsInstalled && desired == StatePresent:
		// the parent feature is installed, we can install the sub feature
		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = c.Namespace
		upgrade.ResetThenReuseValues = true
		// Apply the feature version labels to the release
		upgrade.Labels = map[string]string{
			fmt.Sprintf(FEATURE_VERSION_LABEL, subFeature): subVersion,
		}

		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		release, err := upgrade.RunWithContext(ctx, c.Name, chart, values)
		if err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", c.Name, err)
		}

		// oldConfig and release.Config are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		return !jsonEqual(oldParentConfig, release.Config), nil

	case parentIsInstalled && desired == StateDeleted:
		// the parent feature is installed, we can delete the sub feature

		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = c.Namespace
		upgrade.ResetThenReuseValues = true
		// Remove the label from parent
		upgrade.Labels = map[string]string{
			fmt.Sprintf(FEATURE_VERSION_LABEL, subFeature): "null",
		}

		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		release, err := upgrade.RunWithContext(ctx, c.Name, chart, values)
		if err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", c.Name, err)
		}

		// oldConfig and release.Config are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		return !jsonEqual(oldParentConfig, release.Config), nil

	default:
		// this never happens
		return false, nil

	}
}

func jsonEqual(v1 any, v2 any) bool {
	b1, err1 := json.Marshal(v1)
	b2, err2 := json.Marshal(v2)
	return err1 == nil && err2 == nil && bytes.Equal(b1, b2)
}
