package snap_test

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
)

func TestServiceStartStop(t *testing.T) {
	mockRunner := &MockRunner{}
	s := snap.NewSnap("testdir", "testdir", "testdir", snap.WithCommandRunner(mockRunner.Run))

	tests := []struct {
		name            string
		action          func(ctx context.Context, service string) error
		service         string
		expectedCommand string
	}{
		{
			name:            "StartService",
			action:          s.StartService,
			service:         "test-service",
			expectedCommand: "snapctl start --enable k8s.test-service",
		},
		{
			name:            "StopService",
			action:          s.StopService,
			service:         "test-service",
			expectedCommand: "snapctl stop --disable k8s.test-service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRunner.CalledWithCommand = []string{} // Resetting the commands for each test case
			tc.action(context.Background(), tc.service)
			if lastCmd := mockRunner.CalledWithCommand[0]; lastCmd != tc.expectedCommand {
				t.Fatalf("Expected command %q, but %q was called instead for service %s", tc.expectedCommand, lastCmd, tc.service)
			}
		})
	}
}

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
