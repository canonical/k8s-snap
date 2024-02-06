package types

import apiv1 "github.com/canonical/k8s/api/v1"

// ClusterConfig is the control plane configuration format of the k8s cluster.
// ClusterConfig should attempt to use structured fields wherever possible.
type ClusterConfig struct {
	Cluster      Cluster      `yaml:"cluster"`
	Certificates Certificates `yaml:"certificates"`
	Kubelet      Kubelet      `yaml:"kubelet"`
	K8sDqlite    K8sDqlite    `yaml:"k8s-dqlite"`
	APIServer    APIServer    `yaml:"apiserver"`
}

type Cluster struct {
	CIDR string `yaml:"cidr,omitempty"`
}

type Certificates struct {
	CACert                 string `yaml:"ca-crt,omitempty"`
	CAKey                  string `yaml:"ca-key,omitempty"`
	APIServerToKubeletCert string `yaml:"apiserver-to-kubelet-crt,omitempty"`
	APIServerToKubeletKey  string `yaml:"apiserver-to-kubelet-key,omitempty"`
	K8sDqliteCert          string `yaml:"k8s-dqlite-crt,omitempty"`
	K8sDqliteKey           string `yaml:"k8s-dqlite-key,omitempty"`
	FrontProxyCACert       string `yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey        string `yaml:"front-proxy-ca-key,omitempty"`
}

type Kubelet struct {
	CloudProvider string `yaml:"cloud-provider,omitempty"`
	ClusterDNS    string `yaml:"cluster-dns,omitempty"`
	ClusterDomain string `yaml:"cluster-domain,omitempty"`
}

type APIServer struct {
	SecurePort          int    `yaml:"secure-port,omitempty"`
	AuthorizationMode   string `yaml:"authorization-mode,omitempty"`
	ServiceAccountKey   string `yaml:"service-account-key,omitempty"`
	Datastore           string `yaml:"datastore,omitempty"`
	DatastoreURL        string `yaml:"datastore-url,omitempty"`
	DatastoreCA         string `yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert string `yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  string `yaml:"datastore-client-key,omitempty"`
}

type K8sDqlite struct {
	Port int `yaml:"port,omitempty"`
}

func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Cluster: Cluster{
			CIDR: "10.1.0.0/16",
		},
		APIServer: APIServer{
			SecurePort:        6443,
			AuthorizationMode: "Node,RBAC",
		},
		K8sDqlite: K8sDqlite{
			Port: 9000,
		},
	}
}

// ClusterConfigFromBootstrapConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
func ClusterConfigFromBootstrapConfig(b *apiv1.BootstrapConfig) ClusterConfig {
	return ClusterConfig{
		Cluster: Cluster{
			CIDR: b.ClusterCIDR,
		},
	}
}
