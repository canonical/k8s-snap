package helm

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/provenance"
	"helm.sh/helm/v3/pkg/repo"
)

// InMemoryLoader loads a chart from a byte slice
type InMemoryLoader []byte

// Load loads a chart
func (l InMemoryLoader) Load() (*chart.Chart, error) {
	return LoadBytes(bytes.NewReader(l))
}

// LoadBytes loads from a byte slice.
func LoadBytes(raw *bytes.Reader) (*chart.Chart, error) {
	if err := ensureArchive(raw); err != nil {
		return nil, err
	}

	c, err := loader.LoadArchive(raw)
	if err != nil {
		if err == gzip.ErrHeader {
			return nil, fmt.Errorf("file does not appear to be a valid chart file (details: %s)", err)
		}
	}
	return c, err
}

// isGZipApplication checks whether the achieve is of the application/x-gzip type.
func isGZipApplication(data []byte) bool {
	sig := []byte("\x1F\x8B\x08")
	return bytes.HasPrefix(data, sig)
}

// ensureArchive's job is to return an informative error if the file does not appear to be a gzipped archive.
//
// Sometimes users will provide a values.yaml for an argument where a chart is expected. One common occurrence
// of this is invoking `helm template values.yaml mychart` which would otherwise produce a confusing error
// if we didn't check for this.
func ensureArchive(raw *bytes.Reader) error {
	defer raw.Seek(0, 0) // reset read offset to allow archive loading to proceed.

	// Check the file format to give us a chance to provide the user with more actionable feedback.
	buffer := make([]byte, 512)
	_, err := raw.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("file cannot be read: %s", err)
	}

	// Helm may identify achieve of the application/x-gzip as application/vnd.ms-fontobject.
	// Fix for: https://github.com/helm/helm/issues/12261
	if contentType := http.DetectContentType(buffer); contentType != "application/x-gzip" && !isGZipApplication(buffer) {
		// TODO: Is there a way to reliably test if a file content is YAML? ghodss/yaml accepts a wide
		//       variety of content (Makefile, .zshrc) as valid YAML without errors.
		return fmt.Errorf("file does not appear to be a gzipped archive; got '%s'", contentType)
	}
	return nil
}

func GenerateIndex(charts []types.HelmChart) (*repo.IndexFile, error) {
	indexFile := repo.NewIndexFile()

	for _, chart := range charts {
		ch, err := InMemoryLoader(chart.Contents).Load()
		if err != nil {
			return nil, err
		}

		digest, err := provenance.Digest(bytes.NewReader(chart.Contents))
		if err != nil {
			return nil, err
		}

		if !indexFile.Has(ch.Name(), ch.Metadata.Version) {
			// TODO(berkayoz): Look into baseurl later
			if err := indexFile.MustAdd(ch.Metadata, chart.Name, "", digest); err != nil {
				return nil, errors.Wrapf(err, "failed adding to %s - %s to index", chart.Name, chart.Version)
			}
		}
	}
	indexFile.SortEntries()
	return indexFile, nil
}
