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
func StartControlPlaneServices(ctx context.Context, snap snap.Snap, datastore string) error {
	var services []string
	if datastore == "k8s-dqlite" {
		services = append([]string{"k8s-dqlite"}, controlPlaneServices...)
	}
	for _, service := range services {
		if err := snap.StartService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}

// StopControlPlaneServices stops the control plane services.
// StopControlPlaneServices will return on the first failing service.
func StopControlPlaneServices(ctx context.Context, snap snap.Snap) error {
	var services []string
	services = append([]string{"k8s-dqlite"}, controlPlaneServices...)
	for _, service := range services {
		if err := snap.StopService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}
