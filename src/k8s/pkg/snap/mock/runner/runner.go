package runner

import (
	"context"
	"log"
	"strings"
)

// MockRunner is a mock implementation of CommandRunner.
type MockRunner struct {
	CalledWithCtx     context.Context
	CalledWithCommand []string
	Err               error
	Log               bool
}

// Run is a mock implementation of CommandRunner.
func (m *MockRunner) Run(ctx context.Context, command ...string) error {
	if m.Log {
		log.Printf("mock execute %#v", command)
	}
	m.CalledWithCommand = append(m.CalledWithCommand, strings.Join(command, " "))
	m.CalledWithCtx = ctx
	return m.Err
}
