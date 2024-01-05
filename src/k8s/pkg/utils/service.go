package utils

import (
	"context"
	"fmt"
	"strings"
)

// StartService starts a k8s service. The name can be either prefixed or not.
func StartService(ctx context.Context, name string) error {
	return RunCommand(ctx, "snapctl", "start", serviceName(name))
}

// StopService stops a k8s service. The name can be either prefixed or not.
func StopService(ctx context.Context, name string) error {
	return RunCommand(ctx, "snapctl", "stop", serviceName(name))
}

// serviceName infers the name of the snapctl daemon from the service name.
// if the serviceName is the snap name `k8s` (=referes to all services) it will return it as is.
func serviceName(serviceName string) string {
	if strings.HasPrefix(serviceName, "k8s.") || serviceName == "k8s" {
		return serviceName
	}
	return fmt.Sprintf("k8s.%s", serviceName)
}
