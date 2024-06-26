package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// RunCommand executes a command with a given context.
// RunCommand returns nil if the command completes successfully and the exit code is 0.
func RunCommand(ctx context.Context, command []string, opts ...func(*exec.Cmd)) error {
	var args []string
	if len(command) > 1 {
		args = command[1:]
	}
	cmd := exec.CommandContext(ctx, command[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for _, o := range opts {
		o(cmd)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %v failed with exit code %d: %w", command, cmd.ProcessState.ExitCode(), err)
	}
	return nil
}
