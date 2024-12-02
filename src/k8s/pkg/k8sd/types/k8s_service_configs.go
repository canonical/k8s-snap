package types

import (
	"net"
)

const (
	// Default values for Kubernetes services.
	KubeControllerManagerPort = "10257"
	KubeSchedulerPort         = "10259"
	KubeletPort               = "10250"
	KubeletHealthzPort        = "10248"
	KubeletReadOnlyPort       = "10255"
	KubeProxyHealthzPort      = "10256"
	KubeProxyMetricsPort      = "10249"
)

type K8sServiceConfigs struct {
	ExtraNodeKubeControllerManagerArgs map[string]*string
	ExtraNodeKubeSchedulerArgs         map[string]*string
	ExtraNodeKubeletArgs               map[string]*string
	ExtraNodeKubeProxyArgs             map[string]*string
}

func (s *K8sServiceConfigs) GetKubeControllerManagerPort() string {
	return getConfigOrDefault(s.ExtraNodeKubeControllerManagerArgs, "--secure-port", KubeControllerManagerPort)
}

func (s *K8sServiceConfigs) GetKubeSchedulerPort() string {
	return getConfigOrDefault(s.ExtraNodeKubeSchedulerArgs, "--secure-port", KubeSchedulerPort)
}

func (s *K8sServiceConfigs) GetKubeletPort() string {
	return getConfigOrDefault(s.ExtraNodeKubeletArgs, "--port", KubeletPort)
}

func (s *K8sServiceConfigs) GetKubeletHealthzPort() string {
	return getConfigOrDefault(s.ExtraNodeKubeletArgs, "--healthz-port", KubeletHealthzPort)
}

func (s *K8sServiceConfigs) GetKubeletReadOnlyPort() string {
	return getConfigOrDefault(s.ExtraNodeKubeletArgs, "--read-only-port", KubeletReadOnlyPort)
}

func (s *K8sServiceConfigs) GetKubeProxyHealthzPort() (string, error) {
	address := getConfigOrDefault(s.ExtraNodeKubeProxyArgs, "--healthz-bind-address", "")
	if address == "" {
		return KubeProxyHealthzPort, nil
	}
	_, port, err := net.SplitHostPort(address)
	return port, err
}

func (s *K8sServiceConfigs) GetKubeProxyMetricsPort() (string, error) {
	address := getConfigOrDefault(s.ExtraNodeKubeProxyArgs, "--metrics-bind-address", "")
	if address == "" {
		return KubeProxyMetricsPort, nil
	}
	_, port, err := net.SplitHostPort(address)
	return port, err
}

func getConfigOrDefault(serviceArgs map[string]*string, optionName, defaultValue string) string {
	if serviceArgs == nil {
		return defaultValue
	} else if val, ok := serviceArgs[optionName]; !ok || val == nil {
		return defaultValue
	} else {
		return *val
	}
}
