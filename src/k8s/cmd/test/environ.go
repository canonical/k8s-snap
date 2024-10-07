package test

import (
	"fmt"
	"io"
	"os"
	"strings"

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
}

// DefaultExecutionEnvironment is used to run the CLI.
func DefaultExecutionEnvironment() ExecutionEnvironment {
	var s snap.Snap
	switch os.Getenv("K8SD_RUNTIME_ENVIRONMENT") {
	case "", "snap":
		s = snap.NewSnap(snap.SnapOpts{
			SnapDir:          os.Getenv("SNAP"),
			SnapCommonDir:    os.Getenv("SNAP_COMMON"),
			SnapInstanceName: os.Getenv("SNAP_INSTANCE_NAME"),
		})
	case "pebble":
		s = snap.NewPebble(snap.PebbleOpts{
			SnapDir:       os.Getenv("SNAP"),
			SnapCommonDir: os.Getenv("SNAP_COMMON"),
		})
	default:
		panic(fmt.Sprintf("invalid runtime environment %q", os.Getenv("K8SD_RUNTIME_ENVIRONMENT")))
	}

	return ExecutionEnvironment{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Exit:    os.Exit,
		Environ: os.Environ(),
		Getuid:  os.Getuid,
		Snap:    s,
	}
}

// EnvironWithDefaults returns a copy of the environment.
// EnvironWithDefaults accepts optional key-value pairs to add to the environment (if they are not already set).
func EnvironWithDefaults(environ []string, keyValues ...string) []string {
	if len(keyValues)%2 == 1 {
		panic(fmt.Errorf("key %s does not have a matching value", keyValues[len(keyValues)-1]))
	}

nextKeyValue:
	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]

		for _, val := range environ {
			parts := strings.SplitN(val, "=", 2)
			if parts[0] == key && (len(parts) == 2 || parts[1] != "") {
				continue nextKeyValue
			}
		}

		environ = append(environ, fmt.Sprintf("%s=%s", key, value))
	}

	return environ
}
