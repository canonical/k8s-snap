package snap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
)

// snap implements the Snap interface.
type snap struct {
	snapDir       string
	snapDataDir   string
	snapCommonDir string
}

// NewSnap creates a new interface with the K8s snap.
// NewSnap accepts the $SNAP, $SNAP_DATA and $SNAP_COMMON, directories
func NewSnap(snapDir, snapDataDir, snapCommonDir string, options ...func(s *snap)) Snap {
	s := &snap{
		snapDir:       snapDir,
		snapDataDir:   snapDataDir,
		snapCommonDir: snapCommonDir,
	}

	for _, opt := range options {
		opt(s)
	}
	return s
}

type snapContextKey struct{}

// SnapFromContext extracts the snap instance from the provided context.
// A panic is invoked if there is not snap instance in this context.
func SnapFromContext(ctx context.Context) Snap {
	snap, ok := ctx.Value(snapContextKey{}).(Snap)
	if !ok {
		// This should never happen as the main microcluster state context should contain the snap for k8sd.
		// Thus, panic is fine here to avoid cumbersome and unnecessary error checks on client side.
		panic("There is no snap value in the given context. Make sure that the context is wrapped with snap.ContextWithSnap.")
	}
	return snap
}

// ContextWithSnap adds a snap instance to a given context.
func ContextWithSnap(ctx context.Context, snap Snap) context.Context {
	return context.WithValue(ctx, snapContextKey{}, snap)
}

func (s *snap) Path(parts ...string) string {
	return filepath.Join(append([]string{s.snapDir}, parts...)...)
}

func (s *snap) DataPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapDataDir}, parts...)...)
}
func (s *snap) CommonPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapCommonDir}, parts...)...)
}

// StartService starts a k8s service. The name can be either prefixed or not.
func (s *snap) StartService(ctx context.Context, name string) error {
	return utils.RunCommand(ctx, "snapctl", "start", serviceName(name))
}

// StopService stops a k8s service. The name can be either prefixed or not.
func (s *snap) StopService(ctx context.Context, name string) error {
	return utils.RunCommand(ctx, "snapctl", "stop", serviceName(name))
}

func (s *snap) ReadServiceArguments(serviceName string) (string, error) {
	return utils.ReadFile(s.DataPath("args", serviceName))
}

func (s *snap) WriteServiceArguments(serviceName string, arguments []byte) error {
	return os.WriteFile(s.DataPath("args", serviceName), arguments, 0660)
}

// serviceName infers the name of the snapctl daemon from the service name.
// if the serviceName is the snap name `k8s` (=referes to all services) it will return it as is.
func serviceName(serviceName string) string {
	if strings.HasPrefix(serviceName, "k8s.") || serviceName == "k8s" {
		return serviceName
	}
	return fmt.Sprintf("k8s.%s", serviceName)
}
