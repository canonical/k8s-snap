package snap

import (
	"context"
	"os/exec"
)

// WithCommandRunner configures how shell commands are executed.
func WithCommandRunner(f func(context.Context, []string, ...func(*exec.Cmd)) error) func(s *snap) {
	return func(s *snap) {
		s.runCommand = f
	}
}
