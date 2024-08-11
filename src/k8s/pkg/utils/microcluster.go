package utils

import (
	"encoding/json"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

// MicroclusterMapWithTimeout adds a "timeout" configuration value to the config struct.
// If timeout is zero, the configuration is not affected.
func MicroclusterMapWithTimeout(m map[string]string, timeout time.Duration) map[string]string {
	if timeout == 0 {
		return m
	}
	if m == nil {
		m = make(map[string]string)
	}
	m["_timeout"] = timeout.String()
	return m
}

// MicroclusterTimeoutFromMap returns the configured timeout option from the config struct.
// In case of an invalid or empty value, 0 is returned.
func MicroclusterTimeoutFromMap(m map[string]string) time.Duration {
	if v, ok := m["_timeout"]; !ok {
		return 0
	} else if d, err := time.ParseDuration(v); err != nil {
		return 0
	} else {
		return d
	}
}

// MicroclusterConfigWithBootstrap adds apiv1.BootstrapConfig to the config struct.
func MicroclusterMapWithBootstrapConfig(m map[string]string, bootstrap apiv1.BootstrapConfig) (map[string]string, error) {
	b, err := json.Marshal(bootstrap)
	if err != nil {
		return m, fmt.Errorf("failed to marshal bootstrap config: %w", err)
	}
	if m == nil {
		m = make(map[string]string)
	}
	m["bootstrapConfig"] = string(b)
	return m, nil
}

// MicroclusterBootstrapConfigFromMap returns an apiv1.BootstrapConfig from the config struct.
func MicroclusterBootstrapConfigFromMap(m map[string]string) (apiv1.BootstrapConfig, error) {
	var config apiv1.BootstrapConfig
	if err := json.Unmarshal([]byte(m["bootstrapConfig"]), &config); err != nil {
		return apiv1.BootstrapConfig{}, fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}
	return config, nil
}

// MicroclusterMapWithControlPlaneJoinConfig adds (a JSON formatted) apiv1.ControlPlaneJoinConfig to the config struct.
func MicroclusterMapWithControlPlaneJoinConfig(m map[string]string, controlPlaneJoinConfigJSON string) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m["controlPlaneJoinConfig"] = controlPlaneJoinConfigJSON
	return m
}

// MicroclusterControlPlaneJoinConfigFromMap returns an apiv1.ControlPlaneJoinConfig from the config struct.
func MicroclusterControlPlaneJoinConfigFromMap(m map[string]string) (apiv1.ControlPlaneJoinConfig, error) {
	var config apiv1.ControlPlaneJoinConfig
	if err := json.Unmarshal([]byte(m["controlPlaneJoinConfig"]), &config); err != nil {
		return apiv1.ControlPlaneJoinConfig{}, fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}
	return config, nil
}

// MicroclusterMapWithWorkerJoinConfig adds (a JSON formatted) apiv1.WorkerJoinConfig to the config struct.
func MicroclusterMapWithWorkerJoinConfig(m map[string]string, workerJoinConfigJSON string) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m["workerJoinConfig"] = workerJoinConfigJSON
	return m
}

// MicroclusterWorkerJoinConfigFromMap returns an apiv1.WorkerJoinConfig from the config struct.
func MicroclusterWorkerJoinConfigFromMap(m map[string]string) (apiv1.WorkerJoinConfig, error) {
	var config apiv1.WorkerJoinConfig
	if err := json.Unmarshal([]byte(m["workerJoinConfig"]), &config); err != nil {
		return apiv1.WorkerJoinConfig{}, fmt.Errorf("failed to unmarshal worker join config: %w", err)
	}
	return config, nil
}
