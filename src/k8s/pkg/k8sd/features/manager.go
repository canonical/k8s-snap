package features

import (
	"context"
	"fmt"
	"log"
	"path"
	"reflect"

	"github.com/canonical/k8s/pkg/snap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type helmManager struct {
	restClientGetter func(string) genericclioptions.RESTClientGetter
	manifestsBaseDir string
}

func newHelm(snap snap.Snap) *helmManager {
	return &helmManager{
		restClientGetter: snap.KubernetesRESTClientGetter,
		manifestsBaseDir: snap.ManifestsDir(),
	}
}

func (h *helmManager) newActionConfiguration(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(h.restClientGetter(namespace), namespace, "", log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}
	return actionConfig, nil
}

func (h *helmManager) Apply(ctx context.Context, f feature, desired state, values map[string]any) (bool, error) {
	cfg, err := h.newActionConfiguration(f.namespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	isInstalled := true
	var oldConfig map[string]interface{}

	history := action.NewHistory(cfg)
	history.Max = 1
	releases, err := history.Run(f.name)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			isInstalled = false
		}
		return false, fmt.Errorf("failed to check history of release %s: %w", f.name, err)
	}
	oldConfig = releases[0].Config

	switch {
	case !isInstalled && desired == stateDeleted:
		return false, nil
	case !isInstalled && desired == stateUpgradeOnly:
		return false, fmt.Errorf("cannot upgrade %s as it is not installed", f.name)
	case !isInstalled && desired == statePresent:
		// run an install action
		install := action.NewInstall(cfg)
		install.ReleaseName = f.name
		install.Namespace = f.namespace

		chart, err := loader.Load(path.Join(h.manifestsBaseDir, f.manifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", f.name, err)
		}

		if _, err := install.RunWithContext(ctx, chart, values); err != nil {
			return false, fmt.Errorf("failed to install %s: %w", f.name, err)
		}
		return true, nil
	case isInstalled && desired != stateDeleted:
		// run an upgrade action
		upgrade := action.NewUpgrade(cfg)
		upgrade.Namespace = f.namespace
		upgrade.ReuseValues = true

		chart, err := loader.Load(path.Join(h.manifestsBaseDir, f.manifestPath))
		if err != nil {
			return false, fmt.Errorf("failed to load manifest for %s: %w", f.name, err)
		}

		release, err := upgrade.RunWithContext(ctx, f.name, chart, values)
		if err != nil {
			return false, fmt.Errorf("failed to upgrade %s: %w", f.name, err)
		}

		return !reflect.DeepEqual(oldConfig, release.Config), nil
	case isInstalled && desired == stateDeleted:
		// run an uninstall action
		uninstall := action.NewUninstall(cfg)
		if _, err := uninstall.Run(f.name); err != nil {
			return false, fmt.Errorf("failed to uninstall %s: %w", err)
		}
		return true, nil
	}

	return false, nil
}

var _ Manager = &helmManager{}
