package types

type NodeConfig struct {
	CloudProvider *string
	ClusterDNS    *string
	ClusterDomain *string
}

func NodeConfigFromMap(data map[string]string) NodeConfig {
	nodeConfig := NodeConfig{}

	cloudProvider, ok := data["cloud-provider"]
	if ok {
		nodeConfig.CloudProvider = &cloudProvider
	}

	clusterDNS, ok := data["cluster-dns"]
	if ok {
		nodeConfig.ClusterDNS = &clusterDNS
	}

	clusterDomain, ok := data["cluster-domain"]
	if ok {
		nodeConfig.ClusterDomain = &clusterDomain
	}

	return nodeConfig
}

func MapFromNodeConfig(nodeConfig NodeConfig) map[string]string {
	data := make(map[string]string)

	if nodeConfig.CloudProvider != nil {
		data["cloud-provider"] = *nodeConfig.CloudProvider
	}

	if nodeConfig.ClusterDNS != nil {
		data["cluster-dns"] = *nodeConfig.ClusterDNS
	}

	if nodeConfig.ClusterDomain != nil {
		data["cluster-domain"] = *nodeConfig.ClusterDomain
	}

	return data
}
