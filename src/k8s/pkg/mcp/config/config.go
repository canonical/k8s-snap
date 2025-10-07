package config

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig     `json:"server"`
	Kubernetes KubernetesConfig `json:"kubernetes"`
	Logging    LoggingConfig    `json:"logging"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Address         string `json:"address"` // e.g., ":8080"
	Name            string `json:"name"`
	Version         string `json:"version"`
	ShutdownTimeout int    `json:"shutdown_timeout"` // in seconds
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	KubeConfig string `json:"kubeconfig"`
	InCluster  bool   `json:"in_cluster"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Name:    getEnvWithDefault("SERVER_NAME", "k8s-mcp-server"),
			Version: getEnvWithDefault("SERVER_VERSION", "1.0.0"),
		},
		Kubernetes: KubernetesConfig{
			KubeConfig: getKubeConfigPath(),
			InCluster:  getEnvWithDefault("K8S_IN_CLUSTER", "false") == "true",
		},
		Logging: LoggingConfig{
			Level:  getEnvWithDefault("LOG_LEVEL", "info"),
			Format: getEnvWithDefault("LOG_FORMAT", "text"),
		},
	}

	return cfg, nil
}

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getKubeConfigPath determines the kubeconfig path from environment or default location
func getKubeConfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}
