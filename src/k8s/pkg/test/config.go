package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Configuration is the format of the apiserver proxy endpoints config file.
type Configuration struct {
	Endpoints []string `json:"endpoints"`
}

func loadEndpointsConfig(file string) (Configuration, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg Configuration
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Configuration{}, fmt.Errorf("failed to parse config file %s: %w", file, err)
	}
	sort.Strings(cfg.Endpoints)

	return cfg, nil
}

func WriteEndpointsConfig(endpoints []string, file string) error {
	b, err := json.Marshal(Configuration{Endpoints: endpoints})
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(file, b, 0600); err != nil {
		return fmt.Errorf("failed to write configuration file %s: %w", file, err)
	}
	return nil
}
