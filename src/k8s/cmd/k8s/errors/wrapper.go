package errors

import (
	"errors"
	"fmt"

	v1 "github.com/canonical/k8s/api/v1"
	"github.com/spf13/cobra"
)

var genericErrorMsgs = map[error]string{
	&v1.ErrNotBootstrapped{}: fmt.Sprintln("The cluster has not been initialized yet. Please call:\n\n    sudo k8s bootstrap")
	ErrConnectionFailed: `Unable to connect to the local cluster.`,
}

type ErrorWrapper struct {
	err             error
	CustomErrorMsgs map[error]string
}

// RunE fails to proceed further in case of error resulting in not executing PostRun actions
func (w *ErrorWrapper) Run(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	fmt.Println("test")
	return func(cmd *cobra.Command, args []string) {
		err := f(cmd, args)
		fmt.Println(err)
		w.err = err
	}
}

func (w *ErrorWrapper) TransformToHumanError() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Println("Transform to human error: %v", w.err)
		return w.toHumanError()
	}
}

func (w *ErrorWrapper) toHumanError() error {
	fmt.Println("Error is:", w.err)
	if w.err == nil {
		return nil
	}

	if ew.CustomErrorMsgs != nil {
		// CustomErrorMsgs should have higher prio then generic ones, so we don't merge them.
		for errorType, msg := range ew.CustomErrorMsgs {
			if errors.Is(w.err, errorType) {
				return errors.New(msg)
			}
		}
	}

	for errorType, msg := range genericErrorMsgs {
		if errors.Is(w.err, errorType) {
			return errors.New(msg)
		}
	}
	return w.err
}
