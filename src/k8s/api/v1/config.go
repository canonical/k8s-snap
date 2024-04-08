package v1

import (
	"encoding/json"
	"fmt"
)

type MicroclusterConfig interface {
	ToMicrocluster() (map[string]string, error)
}

// ToMicrocluster implements the conversion to Microcluster for any MicroclusterConfig.
func ToMicrocluster(m MicroclusterConfig, key string) (map[string]string, error) {
	config, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal config %s: %w", key, err)
	}
	return map[string]string{key: string(config)}, nil
}

// ConfigFromMicrocluster parses a Microcluster map[string]string and retrieves the Config structure based on the provided MicroclusterConfig type.
func ConfigFromMicrocluster(m map[string]string, key string, target MicroclusterConfig) error {
	if err := json.Unmarshal([]byte(m[key]), target); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}
	return nil
}
