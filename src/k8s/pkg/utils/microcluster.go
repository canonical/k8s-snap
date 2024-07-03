package utils

import "time"

// MicroclusterConfigWithTimeout adds a "timeout" configuration value to the config struct.
// If timeout is zero, the configuration is not affected.
func MicroclusterConfigWithTimeout(config map[string]string, timeout time.Duration) map[string]string {
	if timeout == 0 {
		return config
	}

	config["_timeout"] = timeout.String()
	return config
}

// MicroclusterTimeoutFromConfig returns the configured timeout option from the config struct.
// If case of an invalid or empty value, 0 is returned.
func MicroclusterTimeoutFromConfig(config map[string]string) time.Duration {
	if v, ok := config["_timeout"]; !ok {
		return 0
	} else if d, err := time.ParseDuration(v); err != nil {
		return 0
	} else {
		return d
	}
}
