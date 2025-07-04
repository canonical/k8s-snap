package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	releasepkg "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// client implements Client using Helm.
type client struct {
	restClientGetter func(string) genericclioptions.RESTClientGetter
	manifestsBaseDir string
	timeout          time.Duration
	// maxHistory specifies the maximum number of historical releases that will
	// be retained, including the most recent release. Values of 0 or less are
	// ignored (meaning no limits are imposed).
	maxHistory int
}

// ensure *client implements Client.
var _ Client = &client{}

// NewClient creates a new client.
func NewClient(manifestsBaseDir string,
	restClientGetter func(string) genericclioptions.RESTClientGetter,
	timeout time.Duration,
	maxHistory int,
) *client {
	return &client{
		restClientGetter: restClientGetter,
		manifestsBaseDir: manifestsBaseDir,
		timeout:          timeout,
		maxHistory:       maxHistory,
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
func (h *client) Apply(ctx context.Context, c InstallableChart, desired State, values map[string]any) (bool, error) {
	log := log.FromContext(ctx).WithName("helm").WithValues("chart", c.Name, "desired", desired)

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
		install.Atomic = true
		install.Timeout = h.timeout
		install.ReleaseName = c.Name
		install.Namespace = c.Namespace
		install.CreateNamespace = true

		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		if _, err := install.RunWithContext(ctx, chart, values); err != nil {
			return false, fmt.Errorf("failed to install %s: %w", c.Name, err)
		}
		return true, nil
	case isInstalled && desired != StateDeleted:
		chart, err := loader.Load(filepath.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		// NOTE(Angelos): oldConfig and values are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		// NOTE(Hue) (KU-3592): We are ignoring the values that are overwritten by the user.
		// The user can change some values in the chart, but we will revert them back upon an upgrade.
		// NOTE(Hue): We clone the values map to avoid modifying the original user provided values.
		clonedValues, err := cloneMap(values)
		if err != nil {
			return false, fmt.Errorf("failed to clone values for %s: %w", c.Name, err)
		}
		mergedValues := chartutil.CoalesceTables(clonedValues, oldConfig)
		sameValues := jsonEqual(oldConfig, mergedValues)
		// NOTE(Hue): For the charts that we manage (e.g. ck-loadbalancer), we need to make
		// sure we bump the version manually. Otherwise, they'll not be applied unless
		// we're lucky and providing different extra values.
		sameVersions := release.Chart.Metadata.Version == chart.Metadata.Version
		switch {
		case sameValues && sameVersions:
			if release.Info.Status == releasepkg.StatusDeployed || release.Info.Status == releasepkg.StatusSuperseded {
				log.Info("no changes detected, skipping upgrade", "status", release.Info.Status)
				return false, nil
			}
			log.Info(fmt.Sprintf("no changes detected, but release status is %q, proceeding with upgrade", release.Info.Status))
		case sameValues && !sameVersions:
			log.Info("chart version changed, upgrading", "oldVersion", release.Chart.Metadata.Version, "newVersion", chart.Metadata.Version)
		case sameVersions && !sameValues:
			log.Info("values changed, upgrading")
		default:
			log.Info("both chart version and values changed, upgrading", "oldVersion", release.Chart.Metadata.Version, "newVersion", chart.Metadata.Version)
		}

		// there is already a release installed, so we must run an upgrade action
		upgrade := action.NewUpgrade(cfg)
		upgrade.Atomic = true
		upgrade.Timeout = h.timeout
		upgrade.Namespace = c.Namespace
		upgrade.ResetThenReuseValues = true
		// NOTE(Hue): We need to set the upgrade.MaxHistory here since it overwrites the
		// cfg.Releases.MaxHistory value.
		upgrade.MaxHistory = h.maxHistory

		if _, err := upgrade.RunWithContext(ctx, c.Name, chart, values); err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", c.Name, err)
		}

		return true, nil
	case isInstalled && desired == StateDeleted:
		// run an uninstall action
		uninstall := action.NewUninstall(cfg)
		uninstall.Timeout = h.timeout
		uninstall.Wait = true
		if _, err := uninstall.Run(c.Name); err != nil {
			return false, fmt.Errorf("failed to uninstall %s: %w", c.Name, err)
		}

		return true, nil
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

// cloneMap creates a deep copy of a map[string]any by marshaling and unmarshaling it.
func cloneMap(m map[string]any) (map[string]any, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}
	var cloned map[string]any
	if err := json.Unmarshal(b, &cloned); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map: %w", err)
	}
	return cloned, nil
}
