package snap

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
)

type PebbleOpts struct {
	SnapDir           string
	SnapCommonDir     string
	RunCommand        func(ctx context.Context, command []string, opts ...func(c *exec.Cmd)) error
	ContainerdBaseDir string
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

	containerdBaseDir := opts.ContainerdBaseDir
	if containerdBaseDir == "" {
		containerdBaseDir = "/"
		if s.Strict() {
			containerdBaseDir = opts.SnapCommonDir
		}
	}
	s.containerdBaseDir = containerdBaseDir

	return s
}

// buildPebbleCommand builds a pebble command with the given subcommand and names.
func (s *pebble) buildPebbleCommand(subcommand string, names []string, extraPebbleArgs ...string) []string {
	return append([]string{filepath.Join(s.snapDir, "bin", "pebble"), subcommand}, append(names, extraPebbleArgs...)...)
}

// StartServices starts a k8s service. The name can be either prefixed or not.
func (s *pebble) StartServices(ctx context.Context, names []string, extraPebbleArgs ...string) error {
	return s.runCommand(ctx, s.buildPebbleCommand("start", names, extraPebbleArgs...))
}

// StopServices stops a k8s service. The name can be either prefixed or not.
func (s *pebble) StopServices(ctx context.Context, names []string, extraPebbleArgs ...string) error {
	return s.runCommand(ctx, s.buildPebbleCommand("stop", names, extraPebbleArgs...))
}

// RestartServices restarts a k8s service. The name can be either prefixed or not.
func (s *pebble) RestartServices(ctx context.Context, names []string, extraPebbleArgs ...string) error {
	return s.runCommand(ctx, s.buildPebbleCommand("restart", names, extraPebbleArgs...))
}

func (s *pebble) Refresh(ctx context.Context, to types.RefreshOpts) (string, error) {
	switch {
	case to.Revision != "":
		return "", fmt.Errorf("pebble does not support refreshing to a revision, only a local path")
	case to.Channel != "":
		return "", fmt.Errorf("pebble does not support refreshing to a channel, only a local path")
	case to.LocalPath != "":
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				log.FromContext(ctx).Info("Refreshing kubernetes snap")
			}
			// replace the "kubernetes" binary with the new source.
			// "cp -f" will replace the binary in case it's currently in use.
			if err := s.runCommand(ctx, []string{"cp", "-f", to.LocalPath, filepath.Join(s.snapDir, "bin", "kubernetes")}); err != nil {
				log.FromContext(ctx).Error(err, "Warning: failed to update the kubernetes binary")
			}
			// restart services if already running.
			services := []string{"kube-apiserver", "kubelet", "kube-controller-manager", "kube-proxy", "kube-scheduler"}
			if err := s.RestartServices(ctx, services); err != nil {
				log.FromContext(ctx).WithValues("services", services).Error(err, "Warning: failed to restart after updating kubernetes binary")
			}
		}()
		return "0", nil
	default:
		return "", fmt.Errorf("empty refresh options")
	}
}

func (s *pebble) RefreshStatus(ctx context.Context, changeID string) (*types.RefreshStatus, error) {
	// pebble does not support refresh status checks
	return &types.RefreshStatus{
		Status: "Done",
		Ready:  true,
		Err:    "",
	}, nil
}

func (s *pebble) Revision(ctx context.Context) (int, error) {
	return 0, nil
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
