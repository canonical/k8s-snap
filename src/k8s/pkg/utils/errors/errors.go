package errors

import "errors"

// DeeplyUnwrapError unwraps an wrapped error.
// DeeplyUnwrapError will return the innermost error for deeply nested errors.
// DeeplyUnwrapError will return the existing error if the error is not wrapped.
func DeeplyUnwrapError(err error) error {
	for {
		cause := errors.Unwrap(err)
		if cause == nil {
			return err
		}
		err = cause
	}
}
