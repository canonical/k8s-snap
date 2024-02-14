package v1

import (
	"github.com/canonical/k8s/pkg/utils/errors"
)

var Errors = []error{
	&ErrNotBootstrapped{},
}

// Server-side errors
type ErrNotBootstrapped struct{}

func (e ErrNotBootstrapped) Error() string {
	return "no such file or directory"
}

func (e ErrNotBootstrapped) Is(err error) bool {
	return errors.DeeplyUnwrapError(err).Error() == e.Error()
}
