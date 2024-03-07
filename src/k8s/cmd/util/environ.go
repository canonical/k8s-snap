package cmdutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/snap"
)

// ExecutionEnvironment wraps everything that is needed for commands to interact with their environment.
type ExecutionEnvironment struct {
	// Stdin is the standard input.
	Stdin io.Reader
	// Stdout is the standard output.
	Stdout io.Writer
	// Stderr is the standard output for errors.
	Stderr io.Writer
	// Exit is used to halt execution with a specific return code.
	Exit func(rc int)
	// Environ is a list of the environment variables, in the form of "KEY=VALUE".
	Environ []string
	// Getuid retrieves the numeric user id of the caller.
	Getuid func() int
	// Snap provides the snap environment for the command.
	Snap snap.Snap
	// Client is used to retrieve a k8sd client.
	Client func(ctx context.Context) (client.Client, error)
}

// DefaultExecutionEnvironment is used to run the CLI.
func DefaultExecutionEnvironment() ExecutionEnvironment {
	snap := snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON"))

	return ExecutionEnvironment{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Exit:    os.Exit,
		Environ: os.Environ(),
		Getuid:  os.Getuid,
		Snap:    snap,
		Client: func(ctx context.Context) (client.Client, error) {
			c, err := client.NewClient(ctx, client.ClusterOpts{
				Snap: snap,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create k8sd client: %w", err)
			}
			return c, nil
		},
	}
}

// EnvWithKeyIfMissing returns a copy of the environment and sets a key if it is not already set.
func EnvWithKeyIfMissing(environ []string, key string, value string) []string {
	for _, val := range environ {
		parts := strings.SplitN(val, "=", 2)
		if parts[0] == key && (len(parts) == 2 || parts[1] != "") {
			return environ
		}
	}
	return append(environ, fmt.Sprintf("%s=%s", key, value))
}
