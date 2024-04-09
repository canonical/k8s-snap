package v1

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

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
func (c *ControlPlaneNodeJoinConfig) ToMicrocluster() (map[string]string, error) {
	config, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control plane join config: %w", err)
	}

	return map[string]string{
		"controlPlaneJoinConfig": string(config),
	}, nil
}

// ToMicrocluster converts a BootstrapConfig to a map[string]string for use in microcluster.
func (w *WorkerNodeJoinConfig) ToMicrocluster() (map[string]string, error) {
	config, err := json.Marshal(w)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control plane join config: %w", err)
	}

	return map[string]string{
		"workerJoinConfig": string(config),
	}, nil
}

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
