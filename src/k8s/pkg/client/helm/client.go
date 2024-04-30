package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

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

func (h *client) newActionConfiguration(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(h.restClientGetter(namespace), namespace, "", log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}
	return actionConfig, nil
}

// Apply implements the Client interface.
func (h *client) Apply(ctx context.Context, c InstallableChart, desired State, values map[string]any) (bool, error) {
	cfg, err := h.newActionConfiguration(c.Namespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	isInstalled := true
	var oldConfig map[string]any

	// get the latest Helm release with the specified name
	get := action.NewGet(cfg)
	release, err := get.Run(c.Name)
	if err != nil {
		if err != driver.ErrReleaseNotFound {
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
		install.ReleaseName = c.Name
		install.Namespace = c.Namespace

		chart, err := loader.Load(path.Join(h.manifestsBaseDir, c.ManifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", c.Name, err)
		}

		if _, err := install.RunWithContext(ctx, chart, values); err != nil {
			return false, fmt.Errorf("failed to install %s: %w", c.Name, err)
		}
		return true, nil
	case isInstalled && desired != StateDeleted:
		// there is already a release installed, so we must run an upgrade action
		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = c.Namespace
		upgrade.ReuseValues = true

		chart, err := loader.Load(path.Join(h.manifestsBaseDir, c.ManifestPath))
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

func jsonEqual(v1 any, v2 any) bool {
	b1, err1 := json.Marshal(v1)
	b2, err2 := json.Marshal(v2)
	return err1 == nil && err2 == nil && bytes.Equal(b1, b2)
}
