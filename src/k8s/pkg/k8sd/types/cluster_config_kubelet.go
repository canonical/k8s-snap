package types

type Kubelet struct {
	CloudProvider *string `json:"cloud-provider,omitempty"`
	ClusterDNS    *string `json:"cluster-dns,omitempty"`
	ClusterDomain *string `json:"cluster-domain,omitempty"`
}

func (c Kubelet) GetCloudProvider() string { return getField(c.CloudProvider) }
func (c Kubelet) GetClusterDNS() string    { return getField(c.ClusterDNS) }
func (c Kubelet) GetClusterDomain() string { return getField(c.ClusterDomain) }
func (c Kubelet) Empty() bool {
	return c.CloudProvider == nil && c.ClusterDNS == nil && c.ClusterDomain == nil
}

// ToConfigMap converts a Kubelet config to a map[string]string to store in a Kubernetes configmap.
func (c Kubelet) ToConfigMap() (map[string]string, error) {
	data := make(map[string]string)

	if v := c.CloudProvider; v != nil {
		data["cloud-provider"] = *v
	}
	if v := c.ClusterDNS; v != nil {
		data["cluster-dns"] = *v
	}
	if v := c.ClusterDomain; v != nil {
		data["cluster-domain"] = *v
	}

	return data, nil
}

// KubeletFromConfigMap parses configmap data into a Kubelet config.
func KubeletFromConfigMap(m map[string]string) (Kubelet, error) {
	var c Kubelet
	if m == nil {
		return c, nil
	}

	if v, ok := m["cloud-provider"]; ok {
		c.CloudProvider = &v
	}
	if v, ok := m["cluster-dns"]; ok {
		c.ClusterDNS = &v
	}
	if v, ok := m["cluster-domain"]; ok {
		c.ClusterDomain = &v
	}

	return c, nil
}
