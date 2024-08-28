package snap

import (
	"fmt"
	"strings"
)

// serviceName infers the name of the snapctl daemon from the service name.
// if the serviceName is the snap name `k8s` (=referes to all services) it will return it as is.
func serviceName(appName string) string {
	if strings.HasPrefix(appName, "k8s.") || appName == "k8s" {
		return appName
	}
	return fmt.Sprintf("k8s.%s", appName)
}

// appName infers the app name from the service.
// It will remove the "k8s." prefix from the name (if any) and return it.
func appName(serviceName string) string {
	return strings.TrimPrefix(serviceName, "k8s.")
}
