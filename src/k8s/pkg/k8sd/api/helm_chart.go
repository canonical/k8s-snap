package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
	"gopkg.in/yaml.v2"
)

type InsertHelmChartRequest struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Contents string `json:"contents"`
}

type InsertHelmChartResponse struct {
}

func (e *Endpoints) postHelmChart(s state.State, r *http.Request) response.Response {
	req := InsertHelmChartRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	contents, err := base64.StdEncoding.DecodeString(req.Contents)
	if err != nil {
		return response.BadRequest(fmt.Errorf("failed to decode contents: %w", err))
	}

	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		err = database.InsertHelmChart(ctx, tx, req.Name, req.Version, contents)
		if err != nil {
			return fmt.Errorf("failed to create worker node token: %w", err)
		}
		return err
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
	}

	return response.SyncResponse(true, &InsertHelmChartResponse{})
}

type GetHelmChartsIndexResponse struct {
	Index string `json:"index"`
}

func (e *Endpoints) getHelmChartsIndex(s state.State, r *http.Request) response.Response {
	var charts []types.HelmChart
	var err error

	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		charts, err = database.GetHelmCharts(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to create worker node token: %w", err)
		}
		return err
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
	}

	indexFile, err := helm.GenerateIndex(charts)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to generate index: %w", err))
	}

	index, err := yaml.Marshal(indexFile)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to marshal index: %w", err))
	}

	return response.ManualResponse(func(w http.ResponseWriter) error {
		w.Header().Set("Content-Type", "application/x-yaml")

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(index); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}

		f, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("ResponseWriter is not type http.Flusher")
		}

		f.Flush()
		return nil
	})
}
