package v1

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type ControlPlaneNodeJoinConfig struct {
	ExtraSANS []string `json:"extra-sans,omitempty" yaml:"extra-sans,omitempty"`

	// Seed certificates for external CA
	FrontProxyClientCert            *string `json:"front-proxy-client-crt,omitempty" yaml:"front-proxy-client-crt,omitempty"`
	FrontProxyClientKey             *string `json:"front-proxy-client-key,omitempty" yaml:"front-proxy-client-key,omitempty"`
	KubeProxyClientCert             *string `json:"kube-proxy-client-crt,omitempty" yaml:"kube-proxy-client-crt,omitempty"`
	KubeProxyClientKey              *string `json:"kube-proxy-client-key,omitempty" yaml:"kube-proxy-client-key,omitempty"`
	KubeSchedulerClientCert         *string `json:"kube-scheduler-client-crt,omitempty" yaml:"kube-scheduler-client-crt,omitempty"`
	KubeSchedulerClientKey          *string `json:"kube-scheduler-client-key,omitempty" yaml:"kube-scheduler-client-key,omitempty"`
	KubeControllerManagerClientCert *string `json:"kube-controller-manager-client-crt,omitempty" yaml:"kube-controller-manager-client-crt,omitempty"`
	KubeControllerManagerClientKey  *string `json:"kube-controller-manager-client-key,omitempty" yaml:"kube-ControllerManager-client-key,omitempty"`

	APIServerCert     *string `json:"apiserver-crt,omitempty" yaml:"apiserver-crt,omitempty"`
	APIServerKey      *string `json:"apiserver-key,omitempty" yaml:"apiserver-key,omitempty"`
	KubeletCert       *string `json:"kubelet-crt,omitempty" yaml:"kubelet-crt,omitempty"`
	KubeletKey        *string `json:"kubelet-key,omitempty" yaml:"kubelet-key,omitempty"`
	KubeletClientCert *string `json:"kubelet-client-crt,omitempty" yaml:"kubelet-client-crt,omitempty"`
	KubeletClientKey  *string `json:"kubelet-client-key,omitempty" yaml:"kubelet-client-key,omitempty"`

	// ExtraNodeConfigFiles will be written to /var/snap/k8s/common/args/conf.d
	ExtraNodeConfigFiles map[string]string `json:"extra-node-config-files,omitempty" yaml:"extra-node-config-files,omitempty"`

	// Extra args to add to individual services (set any arg to null to delete)
	ExtraNodeKubeAPIServerArgs         map[string]*string `json:"extra-node-kube-apiserver-args,omitempty" yaml:"extra-node-kube-apiserver-args,omitempty"`
	ExtraNodeKubeControllerManagerArgs map[string]*string `json:"extra-node-kube-controller-manager-args,omitempty" yaml:"extra-node-kube-controller-manager-args,omitempty"`
	ExtraNodeKubeSchedulerArgs         map[string]*string `json:"extra-node-kube-scheduler-args,omitempty" yaml:"extra-node-kube-scheduler-args,omitempty"`
	ExtraNodeKubeProxyArgs             map[string]*string `json:"extra-node-kube-proxy-args,omitempty" yaml:"extra-node-kube-proxy-args,omitempty"`
	ExtraNodeKubeletArgs               map[string]*string `json:"extra-node-kubelet-args,omitempty" yaml:"extra-node-kubelet-args,omitempty"`
	ExtraNodeContainerdArgs            map[string]*string `json:"extra-node-containerd-args,omitempty" yaml:"extra-node-containerd-args,omitempty"`
	ExtraNodeK8sDqliteArgs             map[string]*string `json:"extra-node-k8s-dqlite-args,omitempty" yaml:"extra-node-k8s-dqlite-args,omitempty"`
}

type WorkerNodeJoinConfig struct {
	KubeletCert         *string `json:"kubelet-crt,omitempty" yaml:"kubelet-crt,omitempty"`
	KubeletKey          *string `json:"kubelet-key,omitempty" yaml:"kubelet-key,omitempty"`
	KubeletClientCert   *string `json:"kubelet-client-crt,omitempty" yaml:"kubelet-client-crt,omitempty"`
	KubeletClientKey    *string `json:"kubelet-client-key,omitempty" yaml:"kubelet-client-key,omitempty"`
	KubeProxyClientCert *string `json:"kube-proxy-client-crt,omitempty" yaml:"kube-proxy-client-crt,omitempty"`
	KubeProxyClientKey  *string `json:"kube-proxy-client-key,omitempty" yaml:"kube-proxy-client-key,omitempty"`

	// ExtraNodeConfigFiles will be written to /var/snap/k8s/common/args/conf.d
	ExtraNodeConfigFiles map[string]string `json:"extra-node-config-files,omitempty" yaml:"extra-node-config-files,omitempty"`

	// Extra args to add to individual services (set any arg to null to delete)
	ExtraNodeKubeProxyArgs         map[string]*string `json:"extra-node-kube-proxy-args,omitempty" yaml:"extra-node-kube-proxy-args,omitempty"`
	ExtraNodeKubeletArgs           map[string]*string `json:"extra-node-kubelet-args,omitempty" yaml:"extra-node-kubelet-args,omitempty"`
	ExtraNodeContainerdArgs        map[string]*string `json:"extra-node-containerd-args,omitempty" yaml:"extra-node-containerd-args,omitempty"`
	ExtraNodeK8sAPIServerProxyArgs map[string]*string `json:"extra-node-k8s-apiserver-proxy-args,omitempty" yaml:"extra-node-k8s-apiserver-proxy-args,omitempty"`
}

func (c *ControlPlaneNodeJoinConfig) GetFrontProxyClientCert() string {
	return getField(c.FrontProxyClientCert)
}
func (c *ControlPlaneNodeJoinConfig) GetFrontProxyClientKey() string {
	return getField(c.FrontProxyClientKey)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeProxyClientCert() string {
	return getField(b.KubeProxyClientCert)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeProxyClientKey() string {
	return getField(b.KubeProxyClientKey)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeSchedulerClientCert() string {
	return getField(b.KubeSchedulerClientCert)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeSchedulerClientKey() string {
	return getField(b.KubeSchedulerClientKey)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeControllerManagerClientCert() string {
	return getField(b.KubeControllerManagerClientCert)
}
func (b *ControlPlaneNodeJoinConfig) GetKubeControllerManagerClientKey() string {
	return getField(b.KubeControllerManagerClientKey)
}
func (c *ControlPlaneNodeJoinConfig) GetAPIServerCert() string { return getField(c.APIServerCert) }
func (c *ControlPlaneNodeJoinConfig) GetAPIServerKey() string  { return getField(c.APIServerKey) }
func (c *ControlPlaneNodeJoinConfig) GetKubeletCert() string   { return getField(c.KubeletCert) }
func (c *ControlPlaneNodeJoinConfig) GetKubeletKey() string    { return getField(c.KubeletKey) }
func (c *ControlPlaneNodeJoinConfig) GetKubeletClientCert() string {
	return getField(c.KubeletClientCert)
}
func (c *ControlPlaneNodeJoinConfig) GetKubeletClientKey() string {
	return getField(c.KubeletClientKey)
}

func (w *WorkerNodeJoinConfig) GetKubeletCert() string       { return getField(w.KubeletCert) }
func (w *WorkerNodeJoinConfig) GetKubeletKey() string        { return getField(w.KubeletKey) }
func (w *WorkerNodeJoinConfig) GetKubeletClientCert() string { return getField(w.KubeletClientCert) }
func (w *WorkerNodeJoinConfig) GetKubeletClientKey() string  { return getField(w.KubeletClientKey) }
func (w *WorkerNodeJoinConfig) GetKubeProxyClientCert() string {
	return getField(w.KubeProxyClientCert)
}
func (w *WorkerNodeJoinConfig) GetKubeProxyClientKey() string { return getField(w.KubeProxyClientKey) }

// WorkerJoinConfigFromMicrocluster parses a microcluster map[string]string and retrieves the WorkerNodeJoinConfig.
func ControlPlaneJoinConfigFromMicrocluster(m map[string]string) (ControlPlaneNodeJoinConfig, error) {
	config := ControlPlaneNodeJoinConfig{}
	if err := yaml.UnmarshalStrict([]byte(m["controlPlaneJoinConfig"]), &config); err != nil {
		return ControlPlaneNodeJoinConfig{}, fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}
	return config, nil
}

// WorkerJoinConfigFromMicrocluster parses a microcluster map[string]string and retrieves the WorkerNodeJoinConfig.
func WorkerJoinConfigFromMicrocluster(m map[string]string) (WorkerNodeJoinConfig, error) {
	config := WorkerNodeJoinConfig{}
	if err := yaml.UnmarshalStrict([]byte(m["workerJoinConfig"]), &config); err != nil {
		return WorkerNodeJoinConfig{}, fmt.Errorf("failed to unmarshal worker join config: %w", err)
	}
	return config, nil
}
