package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// sanitiseMap converts a map with interface{} keys to a map with string keys.
// This is useful for preparing data for use with the Helm client, which requires
// map keys to be strings. Nested maps are also recursively processed to ensure
// all keys are converted to strings.
func SanitiseMap(m map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		switch t := value.(type) {
		case map[interface{}]interface{}:
			result[fmt.Sprint(key)] = SanitiseMap(t)
		default:
			result[fmt.Sprint(key)] = value
		}
	}
	return result
}

// runCommand executes a command with a given context.
// runCommand returns nil if the command completes successfully and the exit code is 0.
func RunCommand(ctx context.Context, command ...string) error {
	var args []string
	if len(command) > 1 {
		args = command[1:]
	}
	cmd := exec.CommandContext(ctx, command[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %v failed with exit code %d: %w", command, cmd.ProcessState.ExitCode(), err)
	}
	return nil
}
