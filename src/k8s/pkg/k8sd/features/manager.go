package features

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// helmManager implements Manager using Helm.
type helmManager struct {
	restClientGetter func(string) genericclioptions.RESTClientGetter
	manifestsBaseDir string
}

// ensure *helmManager implements Manager.
var _ Manager = &helmManager{}

// newHelm creates a new helmManager.
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

// Apply implements the Manager interface.
func (h *helmManager) Apply(ctx context.Context, f Feature, desired state, values map[string]any) (bool, error) {
	cfg, err := h.newActionConfiguration(f.namespace)
	if err != nil {
		return false, fmt.Errorf("failed to create action configuration: %w", err)
	}

	isInstalled := true
	var oldConfig map[string]any

	// get the latest Helm release with the specified name
	get := action.NewGet(cfg)
	release, err := get.Run(f.name)
	if err != nil {
		if err != driver.ErrReleaseNotFound {
			return false, fmt.Errorf("failed to get status of release %s: %w", f.name, err)
		}
		isInstalled = false
	} else {
		// keep the existing release configuration, to check if any changes were made.
		oldConfig = release.Config
	}

	switch {
	case !isInstalled && desired == stateDeleted:
		// no-op
		return false, nil
	case !isInstalled && desired == stateUpgradeOnly:
		// there is no release installed, this is an error
		return false, fmt.Errorf("cannot upgrade %s as it is not installed", f.name)
	case !isInstalled && desired == statePresent:
		// there is no release installed, so we must run an install action
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
		// there is already a release installed, so we must run an install action
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

		// oldConfig and release.Config are the previous and current values. they are compared by checking their respective JSON, as that is good enough for our needs of comparing unstructured map[string]any data.
		return !jsonEqual(oldConfig, release.Config), nil
	case isInstalled && desired == stateDeleted:
		// run an uninstall action
		uninstall := action.NewUninstall(cfg)
		if _, err := uninstall.Run(f.name); err != nil {
			return false, fmt.Errorf("failed to uninstall %s: %w", f.name, err)
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
