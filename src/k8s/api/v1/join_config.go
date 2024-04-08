package v1

type ControlPlaneNodeJoinConfig struct {
	APIServerCert *string `json:"apiserver-crt,omitempty" yaml:"apiserver-crt,omitempty"`
	APIServerKey  *string `json:"apiserver-key,omitempty" yaml:"apiserver-key,omitempty"`
	KubeletCert   *string `json:"kubelet-crt,omitempty" yaml:"kubelet-crt,omitempty"`
	KubeletKey    *string `json:"kubelet-key,omitempty" yaml:"kubelet-key,omitempty"`
}

type WorkerNodeJoinConfig struct {
	KubeletCert *string `json:"kubelet-crt,omitempty" yaml:"kubelet-crt,omitempty"`
	KubeletKey  *string `json:"kubelet-key,omitempty" yaml:"kubelet-key,omitempty"`
}

func (c *ControlPlaneNodeJoinConfig) GetAPIServerCert() string { return getField(c.APIServerCert) }
func (c *ControlPlaneNodeJoinConfig) GetAPIServerKey() string  { return getField(c.APIServerKey) }
func (c *ControlPlaneNodeJoinConfig) GetKubeletCert() string   { return getField(c.KubeletCert) }
func (c *ControlPlaneNodeJoinConfig) GetKubeletKey() string    { return getField(c.KubeletKey) }

func (w *WorkerNodeJoinConfig) GetKubeletCert() string { return getField(w.KubeletCert) }
func (w *WorkerNodeJoinConfig) GetKubeletKey() string  { return getField(w.KubeletKey) }

// ToMicrocluster converts a BootstrapConfig to a map[string]string for use in microcluster.
func (j *ControlPlaneNodeJoinConfig) ToMicrocluster() (map[string]string, error) {
	return ToMicrocluster(j, "joinClusterConfig")
}

// ToMicrocluster converts a BootstrapConfig to a map[string]string for use in microcluster.
func (w *WorkerNodeJoinConfig) ToMicrocluster() (map[string]string, error) {
	return ToMicrocluster(w, "joinClusterConfig")
}
