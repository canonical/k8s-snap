package loader

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/state"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// DatabaseLoader is a helm chart loader that loads charts from the microcluster database.
type databaseLoader struct {
	s state.State
}

// NewDatabaseLoader creates a new database loader.
func NewDatabaseLoader(s state.State) *databaseLoader {
	return &databaseLoader{
		s: s,
	}
}

// Load loads a helm chart from the microcluster database by name and version.
func (l *databaseLoader) Load(ctx context.Context, f helm.InstallableChart) (*chart.Chart, error) {
	var chartEntry *types.HelmChartEntry
	if err := l.s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		chartEntry, err = database.GetHelmChart(ctx, tx, f.Name, f.Version)
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to get helm chart: %w", err)
	}

	chart, err := loader.LoadArchive(bytes.NewReader(chartEntry.Contents))
	if err != nil {
		return nil, fmt.Errorf("failed to load helm chart: %w", err)
	}

	return chart, nil
}
