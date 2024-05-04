package mock

import (
	"context"
	"log"
	"os/exec"
	"strings"
)

type Runner struct {
	CalledWithCtx     context.Context
	CalledWithCommand []string
	Err               error
	Log               bool
}

// Run is a mock implementation of CommandRunner.
func (m *Runner) Run(ctx context.Context, command []string, opts ...func(*exec.Cmd)) error {
	if m.Log {
		log.Printf("mock execute %#v", command)
	}
	m.CalledWithCommand = append(m.CalledWithCommand, strings.Join(command, " "))
	m.CalledWithCtx = ctx
	return m.Err
}
