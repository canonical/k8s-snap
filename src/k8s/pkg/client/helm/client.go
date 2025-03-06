package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// client implements Client using Helm.
type client struct {
	chartLoader      ChartLoader
	restClientGetter func(string) genericclioptions.RESTClientGetter
}

// ensure *client implements Client.
var _ Client = &client{}

// NewClient creates a new client.
func NewClient(restClientGetter func(string) genericclioptions.RESTClientGetter, chartLoader ChartLoader) *client {
	return &client{
		restClientGetter: restClientGetter,
		chartLoader:      chartLoader,
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
	cfg, err := h.newActionConfiguration(ctx, c.InstallNamespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	isInstalled := true
	var oldConfig map[string]any

	// get the latest Helm release with the specified name
	get := action.NewGet(cfg)
	release, err := get.Run(c.InstallName)
	if err != nil {
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			return false, fmt.Errorf("failed to get status of release %s: %w", c.InstallName, err)
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
		return false, fmt.Errorf("cannot upgrade %s as it is not installed", c.InstallName)
	case !isInstalled && desired == StatePresent:
		// there is no release installed, so we must run an install action
		install := action.NewInstall(cfg)
		install.ReleaseName = c.InstallName
		install.Namespace = c.InstallNamespace
		install.CreateNamespace = true

		chart, err := h.chartLoader.Load(ctx, c)
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.InstallName, err)
		}

		if _, err := install.RunWithContext(ctx, chart, values); err != nil {
			return false, fmt.Errorf("failed to install %s: %w", c.InstallName, err)
		}
		return true, nil
	case isInstalled && desired != StateDeleted:
		// there is already a release installed, so we must run an upgrade action
		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = c.InstallNamespace
		upgrade.ResetThenReuseValues = true

		chart, err := h.chartLoader.Load(ctx, c)
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.InstallName, err)
		}

		release, err := upgrade.RunWithContext(ctx, c.InstallName, chart, values)
		if err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", c.InstallName, err)
		}

		// oldConfig and release.Config are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		return !jsonEqual(oldConfig, release.Config), nil
	case isInstalled && desired == StateDeleted:
		// run an uninstall action
		uninstall := action.NewUninstall(cfg)
		if _, err := uninstall.Run(c.InstallName); err != nil {
			return false, fmt.Errorf("failed to uninstall %s: %w", c.InstallName, err)
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
