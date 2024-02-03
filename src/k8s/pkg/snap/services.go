package snap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
)

var (
	// WorkerServices contains all k8s services that run on a worker node except of k8sd.
	WorkerServices = []string{"containerd", "k8s-apiserver-proxy", "kubelet", "kube-proxy"}
	// ControlPlaneServices contains all k8s services that run on a control plane except of k8sd.
	ControlPlaneServices = []string{
		"containerd", "k8s-dqlite", "kube-apiserver",
		"kube-controller-manager", "kube-proxy",
		"kube-scheduler", "kubelet", "k8s-apiserver-proxy",
	}
)

// StartWorkerServices starts the worker services.
// StartWorkerServices will return on the first failing service.
func StartWorkerServices(ctx context.Context, snap Snap) error {
	for _, service := range WorkerServices {
		if err := snap.StartService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}

// StartControlPlaneServices starts the control plane services.
// StartControlPlaneServices will return on the first failing service.
func StartControlPlaneServices(ctx context.Context, snap Snap) error {
	for _, service := range ControlPlaneServices {
		if err := snap.StartService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service, err)
		}
	}
	return nil
}

// StopWorkerServices stors the worker services.
// StopWorkerServices will return on the first failing service.
func StopWorkerServices(ctx context.Context, snap Snap) error {
	for _, service := range WorkerServices {
		if err := snap.StopService(ctx, service); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", service, err)
		}
	}
	return nil
}

// StopControlPlaneServices stops the control plane services.
// StopControlPlaneServices will return on the first failing service.
func StopControlPlaneServices(ctx context.Context, snap Snap) error {
	for _, service := range ControlPlaneServices {
		if err := snap.StopService(ctx, service); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", service, err)
		}
	}
	return nil
}

// GetServiceArgument retrieves the value of a specific argument from the $SNAP_DATA/args/$service file.
// The argument name should include preceding dashes (e.g. "--secure-port").
// If any errors occur, or the argument is not present, an empty string is returned.
func GetServiceArgument(s Snap, serviceName string, argument string) string {
	arguments, err := s.ReadServiceArguments(serviceName)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(arguments, "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		if key, value := utils.ParseArgumentLine(line); key == argument {
			return value
		}
	}
	return ""
}

// UpdateServiceArguments updates the arguments file for a service.
// UpdateServiceArguments is a no-op if updateList and delete are empty.
// updateList is a map of key-value pairs. It will replace the argument with the new value (or just append).
// delete is a list of arguments to remove completely. The argument is removed if present.
// Returns a boolean whether any of the arguments were changed, as well as any errors that may have occured.
func UpdateServiceArguments(s Snap, serviceName string, updateList []map[string]string, delete []string) (bool, error) {
	if updateList == nil {
		updateList = []map[string]string{}
	}
	if delete == nil {
		delete = []string{}
	}

	// If no updates are requested, exit early
	if len(updateList) == 0 && len(delete) == 0 {
		return false, nil
	}

	deleteMap := make(map[string]struct{}, len(delete))
	for _, k := range delete {
		deleteMap[k] = struct{}{}
	}

	updateMap := make(map[string]string, len(updateList))
	for _, update := range updateList {
		for key, value := range update {
			updateMap[key] = value
		}
	}

	arguments, err := s.ReadServiceArguments(serviceName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("failed to read arguments of service %s: %w", serviceName, err)
	}

	changed := false
	existingArguments := make(map[string]struct{}, len(arguments))
	newArguments := make([]string, 0, len(arguments))
	for _, line := range strings.Split(arguments, "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		key, oldValue := utils.ParseArgumentLine(line)
		existingArguments[key] = struct{}{}
		if newValue, ok := updateMap[key]; ok {
			// update argument with new value
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, newValue))
			if oldValue != newValue {
				changed = true
			}
		} else if _, ok := deleteMap[key]; ok {
			// remove argument
			changed = true
			continue
		} else {
			// no change
			newArguments = append(newArguments, line)
		}
	}

	for key, value := range updateMap {
		if _, argExists := existingArguments[key]; !argExists {
			changed = true
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, value))
		}
	}

	if err := s.WriteServiceArguments(serviceName, []byte(strings.Join(newArguments, "\n")+"\n")); err != nil {
		return false, fmt.Errorf("failed to update arguments for service %s: %q", serviceName, err)
	}
	return changed, nil
}
