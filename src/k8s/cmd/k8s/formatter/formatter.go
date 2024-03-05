package formatter

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type Formatter interface {
	// Print formattes the output.
	Print(any) error
}

// New creates a new formatter based on passed type
// Can be "plain", "json", "yaml".
func New(formatterType string, writer io.Writer) (Formatter, error) {
	switch formatterType {
	case "plain":
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

func (p plainFormatter) Print(data any) error {
	_, err := fmt.Fprint(p.writer, data, "\n")
	return err
}

type jsonFormatter struct {
	writer io.Writer
}

func (j jsonFormatter) Print(data any) error {
	encoder := json.NewEncoder(j.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

type yamlFormatter struct {
	writer io.Writer
}

func (y yamlFormatter) Print(data any) error {
	return yaml.NewEncoder(y.writer).Encode(data)
}
