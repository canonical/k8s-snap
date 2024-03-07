package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// MinimumNArgs requires at least N args to be passe.
func MinimumNArgs(env ExecutionEnvironment, n int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			err := fmt.Errorf("requires at least %d arg(s), only received %d", n, len(args))
			cmd.PrintErrf("Error: %v\n%s\n", err, cmd.UsageString())
			env.Exit(1)
			return err
		}
		return nil
	}
}

// MaximumNArgs requires at most N args to be passed.
func MaximumNArgs(env ExecutionEnvironment, n int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			err := fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
			cmd.PrintErrf("Error: %v\n%s\n", err, cmd.UsageString())
			env.Exit(1)
			return err
		}
		return nil
	}
}

// ExactArgs requires exactly N args to be passed.
func ExactArgs(env ExecutionEnvironment, n int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			err := fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
			cmd.PrintErrf("Error: %v\n%s\n", err, cmd.UsageString())
			env.Exit(1)
			return err
		}
		return nil
	}
}
