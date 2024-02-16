package types

import (
	"fmt"
	"net"
	"strings"

	apiv1 "github.com/canonical/k8s/api/v1"
)

// ClusterConfig is the control plane configuration format of the k8s cluster.
// ClusterConfig should attempt to use structured fields wherever possible.
type ClusterConfig struct {
	Network      Network      `yaml:"network"`
	Certificates Certificates `yaml:"certificates"`
	Kubelet      Kubelet      `yaml:"kubelet"`
	K8sDqlite    K8sDqlite    `yaml:"k8s-dqlite"`
	APIServer    APIServer    `yaml:"apiserver"`
}

type Network struct {
	PodCIDR     string `yaml:"pod-cidr,omitempty"`
	ServiceCIDR string `yaml:"svc-cidr,omitempty"`
}

type Certificates struct {
	CACert                     string `yaml:"ca-crt,omitempty"`
	CAKey                      string `yaml:"ca-key,omitempty"`
	APIServerKubeletClientCert string `yaml:"apiserver-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey  string `yaml:"apiserver-kubelet-client-key,omitempty"`
	K8sDqliteCert              string `yaml:"k8s-dqlite-crt,omitempty"`
	K8sDqliteKey               string `yaml:"k8s-dqlite-key,omitempty"`
	FrontProxyCACert           string `yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey            string `yaml:"front-proxy-ca-key,omitempty"`
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

func (c *ClusterConfig) Validate() error {
	clusterCIDRs := strings.Split(c.Network.PodCIDR, ", ")
	if len(clusterCIDRs) != 1 && len(clusterCIDRs) != 2 {
		return fmt.Errorf("invalid number of cluster CIDRs: %d", len(clusterCIDRs))
	}

	for _, cidr := range clusterCIDRs {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
	}

	return nil
}

func (c *ClusterConfig) SetDefaults() {
	if c.Network.PodCIDR == "" {
		c.Network.PodCIDR = "10.1.0.0/16"
	}
	c.Network.ServiceCIDR = "10.152.183.0/24"
	c.APIServer.Datastore = "k8s-dqlite"
	c.APIServer.SecurePort = 6443
	c.APIServer.AuthorizationMode = "Node,RBAC"
	c.K8sDqlite.Port = 9000
}

// ClusterConfigFromBootstrapConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
func ClusterConfigFromBootstrapConfig(b *apiv1.BootstrapConfig) ClusterConfig {
	return ClusterConfig{
		Network: Network{
			PodCIDR: b.ClusterCIDR,
		},
	}
}
