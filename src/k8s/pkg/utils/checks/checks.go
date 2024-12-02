package checks

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
)

// CheckK8sServicePorts verifies that the Kubernetes-related ports are free to be used.
// The ports checked depends on whether a node is a control plane node, or a worker node.
func CheckK8sServicePorts(config types.ClusterConfig, serviceConfigs types.K8sServiceConfigs, isControlPlane bool) error {
	var allErrors []error
	ports := map[string]string{
		// Default values from official Kubernetes documentation.
		"kubelet":           serviceConfigs.GetKubeletPort(),
		"kubelet-healthz":   serviceConfigs.GetKubeletHealthzPort(),
		"kubelet-read-only": serviceConfigs.GetKubeletReadOnlyPort(),
		"k8s-dqlite":        strconv.Itoa(config.Datastore.GetK8sDqlitePort()),
		"loadbalancer":      strconv.Itoa(config.LoadBalancer.GetBGPPeerPort()),
	}

	if port, err := serviceConfigs.GetKubeProxyHealthzPort(); err != nil {
		allErrors = append(allErrors, err)
	} else {
		ports["kube-proxy-healhz"] = port
	}

	if port, err := serviceConfigs.GetKubeProxyMetricsPort(); err != nil {
		allErrors = append(allErrors, err)
	} else {
		ports["kube-proxy-metrics"] = port
	}

	if isControlPlane {
		ports["kube-apiserver"] = strconv.Itoa(config.APIServer.GetSecurePort())
		ports["kube-scheduler"] = serviceConfigs.GetKubeSchedulerPort()
		ports["kube-controller-manager"] = serviceConfigs.GetKubeControllerManagerPort()
	} else {
		ports["kube-apiserver-proxy"] = strconv.Itoa(config.APIServer.GetSecurePort())
	}

	for service, port := range ports {
		if port == "0" {
			// Some ports may be set to 0 in order to disable them. No need to check.
			continue
		}
		if open, err := utils.IsLocalPortOpen(port); err != nil {
			// Could not open port due to error.
			allErrors = append(allErrors, fmt.Errorf("could not check port %s (needed by: %s): %w", port, service, err))
		} else if !open {
			allErrors = append(allErrors, fmt.Errorf("port %s (needed by: %s) is already in use.", port, service))
		}
	}

	return errors.Join(allErrors...)
}
