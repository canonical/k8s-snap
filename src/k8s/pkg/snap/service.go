package snap

import (
	"fmt"
	"strings"
)

// serviceName infers the name of the snapctl daemon from the service name.
// if the serviceName is the snap name `k8s` (=referes to all services) it will return it as is.
func serviceName(serviceName string) string {
	if strings.HasPrefix(serviceName, "k8s.") || serviceName == "k8s" {
		return serviceName
	}
	return fmt.Sprintf("k8s.%s", serviceName)
}

// systemdServiceName infers the name of the systemd service from the service name.
func systemdServiceName(serviceName string) string {
	if strings.HasPrefix(serviceName, "snap.k8s.") {
		return serviceName
	}
	return fmt.Sprintf("snap.k8s.%s", serviceName)
}
