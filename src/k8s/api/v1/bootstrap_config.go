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
	ControlPlaneTaints  []string `json:"control-plane-taints,omitempty" yaml:"control-plane-taints,omitempty"`
	PodCIDR             *string  `json:"pod-cidr,omitempty" yaml:"pod-cidr,omitempty"`
	ServiceCIDR         *string  `json:"service-cidr,omitempty" yaml:"service-cidr,omitempty"`
	DisableRBAC         *bool    `json:"disable-rbac,omitempty" yaml:"disable-rbac,omitempty"`
	SecurePort          *int     `json:"secure-port,omitempty" yaml:"secure-port,omitempty"`
	K8sDqlitePort       *int     `json:"k8s-dqlite-port,omitempty" yaml:"k8s-dqlite-port,omitempty"`
	DatastoreType       *string  `json:"datastore-type,omitempty" yaml:"datastore-type,omitempty"`
	DatastoreServers    []string `json:"datastore-servers,omitempty" yaml:"datastore-servers,omitempty"`
	DatastoreCACert     *string  `json:"datastore-ca-crt,omitempty" yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert *string  `json:"datastore-client-crt,omitempty" yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  *string  `json:"datastore-client-key,omitempty" yaml:"datastore-client-key,omitempty"`

	// Seed configuration for certificates
	ExtraSANs []string `json:"extra-sans,omitempty" yaml:"extra-sans,omitempty"`

	// Seed configuration for external certificates
	CACert                     *string `json:"ca-crt,omitempty" yaml:"ca-crt,omitempty"`
	CAKey                      *string `json:"ca-key,omitempty" yaml:"ca-key,omitempty"`
	ClientCACert               *string `json:"client-ca-crt,omitempty" yaml:"client-ca-crt,omitempty"`
	ClientCAKey                *string `json:"client-ca-key,omitempty" yaml:"client-ca-key,omitempty"`
	FrontProxyCACert           *string `json:"front-proxy-ca-crt,omitempty" yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey            *string `json:"front-proxy-ca-key,omitempty" yaml:"front-proxy-ca-key,omitempty"`
	FrontProxyClientCert       *string `json:"front-proxy-client-crt,omitempty" yaml:"front-proxy-client-crt,omitempty"`
	FrontProxyClientKey        *string `json:"front-proxy-client-key,omitempty" yaml:"front-proxy-client-key,omitempty"`
	APIServerKubeletClientCert *string `json:"apiserver-kubelet-client-crt,omitempty" yaml:"apiserver-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey  *string `json:"apiserver-kubelet-client-key,omitempty" yaml:"apiserver-kubelet-client-key,omitempty"`
	ServiceAccountKey          *string `json:"service-account-key,omitempty" yaml:"service-account-key,omitempty"`

	APIServerCert *string `json:"apiserver-crt,omitempty" yaml:"apiserver-crt,omitempty"`
	APIServerKey  *string `json:"apiserver-key,omitempty" yaml:"apiserver-key,omitempty"`
	KubeletCert   *string `json:"kubelet-crt,omitempty" yaml:"kubelet-crt,omitempty"`
	KubeletKey    *string `json:"kubelet-key,omitempty" yaml:"kubelet-key,omitempty"`
}

func (b *BootstrapConfig) GetDatastoreType() string        { return getField(b.DatastoreType) }
func (b *BootstrapConfig) GetDatastoreCACert() string      { return getField(b.DatastoreCACert) }
func (b *BootstrapConfig) GetDatastoreClientCert() string  { return getField(b.DatastoreClientCert) }
func (b *BootstrapConfig) GetDatastoreClientKey() string   { return getField(b.DatastoreClientKey) }
func (b *BootstrapConfig) GetK8sDqlitePort() int           { return getField(b.K8sDqlitePort) }
func (b *BootstrapConfig) GetCACert() string               { return getField(b.CACert) }
func (b *BootstrapConfig) GetCAKey() string                { return getField(b.CAKey) }
func (b *BootstrapConfig) GetClientCACert() string         { return getField(b.ClientCACert) }
func (b *BootstrapConfig) GetClientCAKey() string          { return getField(b.ClientCAKey) }
func (b *BootstrapConfig) GetFrontProxyCACert() string     { return getField(b.FrontProxyCACert) }
func (b *BootstrapConfig) GetFrontProxyCAKey() string      { return getField(b.FrontProxyCAKey) }
func (b *BootstrapConfig) GetFrontProxyClientCert() string { return getField(b.FrontProxyClientCert) }
func (b *BootstrapConfig) GetFrontProxyClientKey() string  { return getField(b.FrontProxyClientKey) }
func (b *BootstrapConfig) GetAPIServerKubeletClientCert() string {
	return getField(b.APIServerKubeletClientCert)
}
func (b *BootstrapConfig) GetAPIServerKubeletClientKey() string {
	return getField(b.APIServerKubeletClientKey)
}
func (b *BootstrapConfig) GetServiceAccountKey() string { return getField(b.ServiceAccountKey) }
func (b *BootstrapConfig) GetAPIServerCert() string     { return getField(b.APIServerCert) }
func (b *BootstrapConfig) GetAPIServerKey() string      { return getField(b.APIServerKey) }
func (b *BootstrapConfig) GetKubeletCert() string       { return getField(b.KubeletCert) }
func (b *BootstrapConfig) GetKubeletKey() string        { return getField(b.KubeletKey) }

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
