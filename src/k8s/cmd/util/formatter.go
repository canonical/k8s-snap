package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/canonical/k8s/pkg/log"
	"gopkg.in/yaml.v2"
)

type Formatter interface {
	// Print formattes the output.
	Print(any)
}

// New creates a new formatter based on passed type
// Can be "plain", "json", "yaml".
func NewFormatter(formatterType string, writer io.Writer) (Formatter, error) {
	switch formatterType {
	case "", "plain":
		return plainFormatter{writer: writer}, nil
	case "json":
		return jsonFormatter{writer: writer}, nil
	case "yaml":
		return yamlFormatter{writer: writer}, nil
	default:
		return nil, fmt.Errorf("unknown formatter type %q", formatterType)
	}
}

type plainFormatter struct {
	writer io.Writer
}

func (p plainFormatter) Print(data any) {
	if _, err := fmt.Fprint(p.writer, data, "\n"); err != nil {
		log.L().WithCallDepth(1).Error(err, "Failed to format output")
	}
}

type jsonFormatter struct {
	writer io.Writer
}

func (j jsonFormatter) Print(data any) {
	encoder := json.NewEncoder(j.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.L().WithCallDepth(1).Error(err, "Failed to format JSON output")
	}
}

type yamlFormatter struct {
	writer io.Writer
}

func (y yamlFormatter) Print(data any) {
	if err := yaml.NewEncoder(y.writer).Encode(data); err != nil {
		log.L().WithCallDepth(1).Error(err, "Failed to format YAML output")
	}
}
