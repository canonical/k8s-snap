package log

import (
	"context"

	"github.com/go-logr/logr"
)

// NewContext returns a new context that contains the logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return logr.NewContext(ctx, logger)
}

// FromContext returns the logger associated with the context.
// FromContext returns the default log.L() logger if there is no logger on the context.
func FromContext(ctx context.Context) Logger {
	if logger, err := logr.FromContext(ctx); err == nil {
		return logger
	}
	return L()
}
