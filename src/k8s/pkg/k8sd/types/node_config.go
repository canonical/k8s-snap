package types

type NodeConfig struct {
	CloudProvider string `mapstructure:"cloud-provider,omitempty"`
	ClusterDNS    string `mapstructure:"cluster-dns,omitempty"`
	ClusterDomain string `mapstructure:"cluster-domain,omitempty"`
}
