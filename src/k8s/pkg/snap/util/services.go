package snaputil

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

var (
	// WorkerServices contains all k8s services that run on a worker node except of k8sd.
	workerServices = []string{
		"containerd",
		"k8s-apiserver-proxy",
		"kubelet",
		"kube-proxy",
	}
	// ControlPlaneServices contains all k8s services that run on a control plane except of k8sd.
	controlPlaneServices = []string{
		"containerd",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-proxy",
		"kube-scheduler",
		"kubelet",
	}
)

// StartWorkerServices starts the worker services.
// StartWorkerServices will return on the first failing service.
func StartWorkerServices(ctx context.Context, snap snap.Snap) error {
	for _, service := range workerServices {
		if err := snap.StartService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}

// StartControlPlaneServices starts the control plane services.
// StartControlPlaneServices will return on the first failing service.
func StartControlPlaneServices(ctx context.Context, snap snap.Snap) error {
	for _, service := range controlPlaneServices {
		if err := snap.StartService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}

// StartK8sDBService starts the managed datastore service.
func StartK8sDBService(ctx context.Context, snap snap.Snap) error {
	if err := snap.StartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to start service %s: %w", "k8s-dqlite", err)
	}
	return nil
}

// StopControlPlaneServices stops the control plane services.
// StopControlPlaneServices will return on the first failing service.
func StopControlPlaneServices(ctx context.Context, snap snap.Snap) error {
	for _, service := range controlPlaneServices {
		if err := snap.StopService(ctx, service); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", service, err)
		}
	}
	return nil
}

// StopK8sDBService stops the managed datastore service.
func StopK8sDBService(ctx context.Context, snap snap.Snap) error {
	if err := snap.StopService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", "k8s-dqlite", err)
	}
	return nil
}

// ServiceArgsFromMap processes a map of string pointers and categorizes them into update and delete lists.
// - If the value pointer is nil, it adds the argument name to the delete list.
// - If the value pointer is not nil, it adds the argument and its value to the update map.
func ServiceArgsFromMap(args map[string]*string) (map[string]string, []string) {
	updateArgs := make(map[string]string)
	deleteArgs := make([]string, 0)

	for arg, val := range args {
		if val == nil {
			deleteArgs = append(deleteArgs, arg)
		} else {
			updateArgs[arg] = *val
		}
	}
	return updateArgs, deleteArgs
}
