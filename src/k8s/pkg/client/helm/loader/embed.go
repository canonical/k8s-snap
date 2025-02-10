package loader

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// embedLoader is a helm chart loader that loads charts from the embedded filesystem.
type embedLoader struct {
	chartFS *embed.FS
}

// NewEmbedLoader creates a new embed loader.
func NewEmbedLoader(chartFS *embed.FS) *embedLoader {
	return &embedLoader{
		chartFS: chartFS,
	}
}

// Load loads a helm chart from the filesystem by name and version.
func (l *embedLoader) Load(_ context.Context, f helm.InstallableChart) (*chart.Chart, error) {
	chartBytes, err := l.chartFS.ReadFile(filepath.Join("charts", fmt.Sprintf("%s-%s.tgz", f.Name, f.Version)))
	if err != nil {
		return nil, fmt.Errorf("failed to read helm chart: %w", err)
	}

	chart, err := loader.LoadArchive(bytes.NewReader(chartBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to load helm chart: %w", err)
	}

	return chart, nil
}
