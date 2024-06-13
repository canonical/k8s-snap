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
	ControlPlaneTaints []string `json:"control-plane-taints,omitempty" yaml:"control-plane-taints,omitempty"`
	PodCIDR            *string  `json:"pod-cidr,omitempty" yaml:"pod-cidr,omitempty"`
	ServiceCIDR        *string  `json:"service-cidr,omitempty" yaml:"service-cidr,omitempty"`
	DisableRBAC        *bool    `json:"disable-rbac,omitempty" yaml:"disable-rbac,omitempty"`
	SecurePort         *int     `json:"secure-port,omitempty" yaml:"secure-port,omitempty"`

	K8sDqlitePort             *int     `json:"k8s-dqlite-port,omitempty" yaml:"k8s-dqlite-port,omitempty"`
	DatastoreType             *string  `json:"datastore-type,omitempty" yaml:"datastore-type,omitempty"`
	DatastoreServers          []string `json:"datastore-servers,omitempty" yaml:"datastore-servers,omitempty"`
	DatastoreCACert           *string  `json:"datastore-ca-crt,omitempty" yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert       *string  `json:"datastore-client-crt,omitempty" yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey        *string  `json:"datastore-client-key,omitempty" yaml:"datastore-client-key,omitempty"`
	DatastoreEmbeddedPort     *int     `json:"datastore-embedded-port,omitempty" yaml:"datastore-embedded-port,omitempty"`
	DatastoreEmbeddedPeerPort *int     `json:"datastore-embedded-peer-port,omitempty" yaml:"datastore-embedded-peer-port,omitempty"`

	// Seed configuration for certificates
	ExtraSANs []string `json:"extra-sans,omitempty" yaml:"extra-sans,omitempty"`

	// Seed configuration for external certificates (cluster-wide)
	CACert                          *string `json:"ca-crt,omitempty" yaml:"ca-crt,omitempty"`
	CAKey                           *string `json:"ca-key,omitempty" yaml:"ca-key,omitempty"`
	ClientCACert                    *string `json:"client-ca-crt,omitempty" yaml:"client-ca-crt,omitempty"`
	ClientCAKey                     *string `json:"client-ca-key,omitempty" yaml:"client-ca-key,omitempty"`
	FrontProxyCACert                *string `json:"front-proxy-ca-crt,omitempty" yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey                 *string `json:"front-proxy-ca-key,omitempty" yaml:"front-proxy-ca-key,omitempty"`
	FrontProxyClientCert            *string `json:"front-proxy-client-crt,omitempty" yaml:"front-proxy-client-crt,omitempty"`
	FrontProxyClientKey             *string `json:"front-proxy-client-key,omitempty" yaml:"front-proxy-client-key,omitempty"`
	APIServerKubeletClientCert      *string `json:"apiserver-kubelet-client-crt,omitempty" yaml:"apiserver-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey       *string `json:"apiserver-kubelet-client-key,omitempty" yaml:"apiserver-kubelet-client-key,omitempty"`
	AdminClientCert                 *string `json:"admin-client-crt,omitempty" yaml:"admin-client-crt,omitempty"`
	AdminClientKey                  *string `json:"admin-client-key,omitempty" yaml:"admin-client-key,omitempty"`
	KubeProxyClientCert             *string `json:"kube-proxy-client-crt,omitempty" yaml:"kube-proxy-client-crt,omitempty"`
	KubeProxyClientKey              *string `json:"kube-proxy-client-key,omitempty" yaml:"kube-proxy-client-key,omitempty"`
	KubeSchedulerClientCert         *string `json:"kube-scheduler-client-crt,omitempty" yaml:"kube-scheduler-client-crt,omitempty"`
	KubeSchedulerClientKey          *string `json:"kube-scheduler-client-key,omitempty" yaml:"kube-scheduler-client-key,omitempty"`
	KubeControllerManagerClientCert *string `json:"kube-controller-manager-client-crt,omitempty" yaml:"kube-controller-manager-client-crt,omitempty"`
	KubeControllerManagerClientKey  *string `json:"kube-controller-manager-client-key,omitempty" yaml:"kube-ControllerManager-client-key,omitempty"`
	ServiceAccountKey               *string `json:"service-account-key,omitempty" yaml:"service-account-key,omitempty"`

	// Seed configuration for embedded datastore
	EmbeddedCACert              *string `json:"embedded-ca-crt,omitempty" yaml:"embedded-ca-crt,omitempty"`
	EmbeddedCAKey               *string `json:"embedded-ca-key,omitempty" yaml:"embedded-ca-key,omitempty"`
	EmbeddedServerCert          *string `json:"embedded-server-crt,omitempty" yaml:"embedded-server-crt,omitempty"`
	EmbeddedServerKey           *string `json:"embedded-server-key,omitempty" yaml:"embedded-server-key,omitempty"`
	EmbeddedServerPeerCert      *string `json:"embedded-peer-crt,omitempty" yaml:"embedded-peer-crt,omitempty"`
	EmbeddedServerPeerKey       *string `json:"embedded-peer-key,omitempty" yaml:"embedded-peer-key,omitempty"`
	EmbeddedAPIServerClientCert *string `json:"embedded-apiserver-client-crt,omitempty" yaml:"embedded-apiserver-client-crt,omitempty"`
	EmbeddedAPIServerClientKey  *string `json:"embedded-apiserver-client-key,omitempty" yaml:"embedded-apiserver-client-key,omitempty"`

	// Seed configuration for external certificates (node-specific)
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

func (b *BootstrapConfig) GetDatastoreType() string       { return getField(b.DatastoreType) }
func (b *BootstrapConfig) GetDatastoreCACert() string     { return getField(b.DatastoreCACert) }
func (b *BootstrapConfig) GetDatastoreClientCert() string { return getField(b.DatastoreClientCert) }
func (b *BootstrapConfig) GetDatastoreClientKey() string  { return getField(b.DatastoreClientKey) }
func (b *BootstrapConfig) GetDatastoreEmbeddedPort() int  { return getField(b.DatastoreEmbeddedPort) }
func (b *BootstrapConfig) GetDatastoreEmbeddedPeerPort() int {
	return getField(b.DatastoreEmbeddedPeerPort)
}
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
func (b *BootstrapConfig) GetAdminClientCert() string     { return getField(b.AdminClientCert) }
func (b *BootstrapConfig) GetAdminClientKey() string      { return getField(b.AdminClientKey) }
func (b *BootstrapConfig) GetKubeProxyClientCert() string { return getField(b.KubeProxyClientCert) }
func (b *BootstrapConfig) GetKubeProxyClientKey() string  { return getField(b.KubeProxyClientKey) }
func (b *BootstrapConfig) GetKubeSchedulerClientCert() string {
	return getField(b.KubeSchedulerClientCert)
}
func (b *BootstrapConfig) GetKubeSchedulerClientKey() string {
	return getField(b.KubeSchedulerClientKey)
}
func (b *BootstrapConfig) GetKubeControllerManagerClientCert() string {
	return getField(b.KubeControllerManagerClientCert)
}
func (b *BootstrapConfig) GetKubeControllerManagerClientKey() string {
	return getField(b.KubeControllerManagerClientKey)
}
func (b *BootstrapConfig) GetServiceAccountKey() string  { return getField(b.ServiceAccountKey) }
func (b *BootstrapConfig) GetEmbeddedCACert() string     { return getField(b.EmbeddedCACert) }
func (b *BootstrapConfig) GetEmbeddedCAKey() string      { return getField(b.EmbeddedCAKey) }
func (b *BootstrapConfig) GetEmbeddedServerCert() string { return getField(b.EmbeddedServerCert) }
func (b *BootstrapConfig) GetEmbeddedServerKey() string  { return getField(b.EmbeddedServerKey) }
func (b *BootstrapConfig) GetEmbeddedServerPeerCert() string {
	return getField(b.EmbeddedServerPeerCert)
}
func (b *BootstrapConfig) GetEmbeddedServerPeerKey() string { return getField(b.EmbeddedServerPeerKey) }
func (b *BootstrapConfig) GetEmbeddedAPIServerClientCert() string {
	return getField(b.EmbeddedAPIServerClientCert)
}
func (b *BootstrapConfig) GetEmbeddedAPIServerClientKey() string {
	return getField(b.EmbeddedAPIServerClientKey)
}
func (b *BootstrapConfig) GetAPIServerCert() string     { return getField(b.APIServerCert) }
func (b *BootstrapConfig) GetAPIServerKey() string      { return getField(b.APIServerKey) }
func (b *BootstrapConfig) GetKubeletCert() string       { return getField(b.KubeletCert) }
func (b *BootstrapConfig) GetKubeletKey() string        { return getField(b.KubeletKey) }
func (b *BootstrapConfig) GetKubeletClientCert() string { return getField(b.KubeletClientCert) }
func (b *BootstrapConfig) GetKubeletClientKey() string  { return getField(b.KubeletClientKey) }

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
