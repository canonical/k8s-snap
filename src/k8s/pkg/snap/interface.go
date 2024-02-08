package snap

import "context"

// Snap is how k8s interacts with the snap.
type Snap interface {
	// IsStrict returns true if the snap is installed with strict confinement.
	IsStrict() bool
	// ReadServiceArguments reads the arguments file for a particular service.
	ReadServiceArguments(serviceName string) (string, error)
	// WriteServiceArguments updates the arguments file a particular service.
	WriteServiceArguments(serviceName string, b []byte) error

	// StartService starts a k8s service.
	StartService(ctx context.Context, serviceName string) error
	// StopService stops a k8s service.
	StopService(ctx context.Context, serviceName string) error
	// RestartService restarts a k8s service.
	RestartService(ctx context.Context, serviceName string) error

	// Path concenates any passed path parts with the $SNAP path
	Path(parts ...string) string
	// DataPath concenates any passed path parts with the $SNAP_DATA path
	DataPath(parts ...string) string
	// CommonPath concenates any passed path parts with the $SNAP_COMMON path
	CommonPath(parts ...string) string
	// WorkerNodeLockFile returns the path to the lockfile that marks a node as worker.
	WorkerNodeLockFile() string
}
