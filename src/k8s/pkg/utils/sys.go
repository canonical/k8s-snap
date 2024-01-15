package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// runCommand executes a command with a given context.
// runCommand returns nil if the command completes successfully and the exit code is 0.
func RunCommand(ctx context.Context, command ...string) error {
	return RunCommandWithCustomWriter(ctx, os.Stdout, os.Stderr, command...)
}

// RunCommandWithCustomWriter executes a command with a given context and allows to set a custom writer
// RunCommandWithCustomWriter returns nil if the command completes successfully and the exit code is 0.
func RunCommandWithCustomWriter(ctx context.Context, stdout io.Writer, stderr io.Writer, command ...string) error {
	var args []string
	if len(command) > 1 {
		args = command[1:]
	}
	cmd := exec.CommandContext(ctx, command[0], args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %v failed with exit code %d: %w", command, cmd.ProcessState.ExitCode(), err)
	}
	return nil
}
