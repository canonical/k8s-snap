package snap

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/canonical/k8s/pkg/utils"
)

type PebbleOpts struct {
	SnapDir       string
	SnapCommonDir string
	RunCommand    func(ctx context.Context, command []string, opts ...func(c *exec.Cmd)) error
}

// pebble implements the Snap interface.
// pebble is the same as snap, but uses pebble for managing services, and disables snapctl.
type pebble struct {
	snap
}

// NewPebble creates a new interface with the K8s snap.
func NewPebble(opts PebbleOpts) *pebble {
	runCommand := utils.RunCommand
	if opts.RunCommand != nil {
		runCommand = opts.RunCommand
	}
	s := &pebble{
		snap: snap{
			snapDir:       opts.SnapDir,
			snapCommonDir: opts.SnapCommonDir,
			runCommand:    runCommand,
		},
	}

	return s
}

// StartService starts a k8s service. The name can be either prefixed or not.
func (s *pebble) StartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{filepath.Join(s.snapDir, "bin", "pebble"), "start", name})
}

// StopService stops a k8s service. The name can be either prefixed or not.
func (s *pebble) StopService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{filepath.Join(s.snapDir, "bin", "pebble"), "stop", name})
}

// RestartService restarts a k8s service. The name can be either prefixed or not.
func (s *pebble) RestartService(ctx context.Context, name string) error {
	return s.runCommand(ctx, []string{filepath.Join(s.snapDir, "bin", "pebble"), "restart", name})
}

func (s *pebble) Strict() bool {
	return false
}

func (s *pebble) OnLXD(ctx context.Context) (bool, error) {
	return true, nil
}

func (s *pebble) SnapctlGet(ctx context.Context, args ...string) ([]byte, error) {
	return []byte(`{"meta": {"apiVersion": "1.30", "orb": "none"}`), nil
}

func (s *pebble) SnapctlSet(ctx context.Context, args ...string) error {
	return nil
}

var _ Snap = &pebble{}
