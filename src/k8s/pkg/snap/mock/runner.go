package mock

import (
	"context"
	"os/exec"
	"strings"
)

type Runner struct {
	CalledWithCtx     context.Context
	CalledWithCommand []string
	Err               error
	Log               bool
	RunOutput         string
}

// Run is a mock implementation of CommandRunner.
func (m *Runner) Run(ctx context.Context, command []string, opts ...func(*exec.Cmd)) error {
	m.CalledWithCommand = append(m.CalledWithCommand, strings.Join(command, " "))
	m.CalledWithCtx = ctx

	// In some cases, it is expected to get the command's stdout.
	cmd := &exec.Cmd{}
	for _, o := range opts {
                o(cmd)
        }
	if m.RunOutput != "" {
		cmd.Stdout.Write([]byte(m.RunOutput))
	}

	return m.Err
}
