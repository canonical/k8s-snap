package cmdutil

import (
	"context"
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
func NewFormatter(formatterType string, writer io.Writer) (Formatter, error) {
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

type formatterContextKey struct{}

func ContextWithFormatter(ctx context.Context, formatter Formatter) context.Context {
	return context.WithValue(ctx, formatterContextKey{}, formatter)
}

func FormatterFromContext(ctx context.Context) Formatter {
	formatter, ok := ctx.Value(formatterContextKey{}).(Formatter)
	if !ok {
		// This should never happen as the main microcluster state context should contain the snap for k8sd.
		// Thus, panic is fine here to avoid cumbersome and unnecessary error checks on client side.
		panic("There is no formatter value in the given context. Make sure that the context is wrapped with cmdutil.ContextWithFormatter.")
	}
	return formatter
}
