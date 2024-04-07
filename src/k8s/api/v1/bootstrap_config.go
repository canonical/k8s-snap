package v1

import (
	"encoding/json"
	"fmt"
)

// BootstrapConfig is used to seed cluster configuration when bootstrapping a new cluster.
type BootstrapConfig struct {
	// ClusterConfig
	ClusterConfig UserFacingClusterConfig `json:"cluster-config,omitempty" yaml:"cluster-config,omitempty"`

	// Seed configuration for the control plane (flat on purpose). Empty values are ignored
	PodCIDR             *string  `json:"pod-cidr,omitempty" yaml:"pod-cidr,omitempty"`
	ServiceCIDR         *string  `json:"service-cidr,omitempty" yaml:"service-cidr,omitempty"`
	DisableRBAC         *bool    `json:"disable-rbac,omitempty" yaml:"disable-rbac,omitempty"`
	SecurePort          *int     `json:"secure-port,omitempty" yaml:"secure-port,omitempty"`
	CloudProvider       *string  `json:"cloud-provider,omitempty" yaml:"cloud-provider,omitempty"`
	K8sDqlitePort       *int     `json:"k8s-dqlite-port,omitempty" yaml:"k8s-dqlite-port,omitempty"`
	DatastoreType       *string  `json:"datastore-type,omitempty" yaml:"datastore-type,omitempty"`
	DatastoreServers    []string `json:"datastore-servers,omitempty" yaml:"datastore-servers,omitempty"`
	DatastoreCACert     *string  `json:"datastore-ca-crt,omitempty" yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert *string  `json:"datastore-client-crt,omitempty" yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  *string  `json:"datastore-client-key,omitempty" yaml:"datastore-client-key,omitempty"`

	// Seed configuration for certificates
	ExtraSANs []string `json:"extra-sans,omitempty" yaml:"extra-sans,omitempty"`
}

func (b *BootstrapConfig) GetDatastoreType() string { return getField(b.DatastoreType) }

// ToMicrocluster converts a BootstrapConfig to a map[string]string for use in microcluster.
func (b *BootstrapConfig) ToMicrocluster() (map[string]string, error) {
	config, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bootstrap config: %w", err)
	}

	return map[string]string{
		"bootstrapConfig": string(config),
	}, nil
}

// BootstrapConfigFromMicrocluster parses a microcluster map[string]string and retrieves the BootstrapConfig.
func BootstrapConfigFromMicrocluster(m map[string]string) (BootstrapConfig, error) {
	config := BootstrapConfig{}
	if err := json.Unmarshal([]byte(m["bootstrapConfig"]), &config); err != nil {
		return BootstrapConfig{}, fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}
	return config, nil
}
