package types

import (
	"fmt"
	"net"
)

// MergeClusterConfig applies updates from non-empty values of the new ClusterConfig to an existing one.
// MergeClusterConfig will return an error if we try to update a config that must not be updated. once such an operation is implemented in the future, we can allow the change here.
// MergeClusterConfig will create a new ClusterConfig object to avoid mutating the existing config objects.
func MergeClusterConfig(existing ClusterConfig, new ClusterConfig) (ClusterConfig, error) {
	var (
		config ClusterConfig
		err    error
	)

	// update string fields
	for _, i := range []struct {
		name        string
		val         **string
		old         *string
		new         *string
		allowChange bool
	}{
		// certificates
		{name: "CA certificate", val: &config.Certificates.CACert, old: existing.Certificates.CACert, new: new.Certificates.CACert},
		{name: "CA key", val: &config.Certificates.CAKey, old: existing.Certificates.CAKey, new: new.Certificates.CAKey},
		{name: "apiserver-kubelet-client certificate", val: &config.Certificates.APIServerKubeletClientCert, old: existing.Certificates.APIServerKubeletClientCert, new: new.Certificates.APIServerKubeletClientCert, allowChange: true},
		{name: "apiserver-kubelet-client key", val: &config.Certificates.APIServerKubeletClientKey, old: existing.Certificates.APIServerKubeletClientKey, new: new.Certificates.APIServerKubeletClientKey, allowChange: true},
		{name: "front proxy CA certificate", val: &config.Certificates.FrontProxyCACert, old: existing.Certificates.FrontProxyCACert, new: new.Certificates.FrontProxyCACert},
		{name: "front proxy CA key", val: &config.Certificates.FrontProxyCAKey, old: existing.Certificates.FrontProxyCAKey, new: new.Certificates.FrontProxyCAKey},
		{name: "service account key", val: &config.Certificates.ServiceAccountKey, old: existing.Certificates.ServiceAccountKey, new: new.Certificates.ServiceAccountKey},
		// datastore
		{name: "datastore type", val: &config.Datastore.Type, old: existing.Datastore.Type, new: new.Datastore.Type},
		{name: "k8s-dqlite certificate", val: &config.Datastore.K8sDqliteCert, old: existing.Datastore.K8sDqliteCert, new: new.Datastore.K8sDqliteCert},
		{name: "k8s-dqlite key", val: &config.Datastore.K8sDqliteKey, old: existing.Datastore.K8sDqliteKey, new: new.Datastore.K8sDqliteKey},
		{name: "external datastore URL", val: &config.Datastore.ExternalURL, old: existing.Datastore.ExternalURL, new: new.Datastore.ExternalURL, allowChange: true},
		{name: "external datastore CA certificate", val: &config.Datastore.ExternalCACert, old: existing.Datastore.ExternalCACert, new: new.Datastore.ExternalCACert, allowChange: true},
		{name: "external datastore client certificate", val: &config.Datastore.ExternalClientCert, old: existing.Datastore.ExternalClientCert, new: new.Datastore.ExternalClientCert, allowChange: true},
		{name: "external datastore client key", val: &config.Datastore.ExternalClientKey, old: existing.Datastore.ExternalClientKey, new: new.Datastore.ExternalClientKey, allowChange: true},
		// network
		{name: "pod CIDR", val: &config.Network.PodCIDR, old: existing.Network.PodCIDR, new: new.Network.PodCIDR},
		{name: "service CIDR", val: &config.Network.ServiceCIDR, old: existing.Network.ServiceCIDR, new: new.Network.ServiceCIDR},
		// apiserver
		{name: "kube-apiserver authorization mode", val: &config.APIServer.AuthorizationMode, old: existing.APIServer.AuthorizationMode, new: new.APIServer.AuthorizationMode, allowChange: true},
		// kubelet
		{name: "kubelet cluster DNS", val: &config.Kubelet.ClusterDNS, old: existing.Kubelet.ClusterDNS, new: new.Kubelet.ClusterDNS, allowChange: true},
		{name: "kubelet cluster domain", val: &config.Kubelet.ClusterDomain, old: existing.Kubelet.ClusterDomain, new: new.Kubelet.ClusterDomain, allowChange: true},
		{name: "kubelet cloud provider", val: &config.Kubelet.CloudProvider, old: existing.Kubelet.CloudProvider, new: new.Kubelet.CloudProvider, allowChange: true},
		// ingress
		{name: "ingress default TLS secret", val: &config.Ingress.DefaultTLSSecret, old: existing.Ingress.DefaultTLSSecret, new: new.Ingress.DefaultTLSSecret, allowChange: true},
		// load balancer
		{name: "load balancer BGP peer address", val: &config.LoadBalancer.BGPPeerAddress, old: existing.LoadBalancer.BGPPeerAddress, new: new.LoadBalancer.BGPPeerAddress, allowChange: true},
		// local storage
		{name: "local storage path", val: &config.LocalStorage.LocalPath, old: existing.LocalStorage.LocalPath, new: new.LocalStorage.LocalPath, allowChange: !existing.LocalStorage.GetEnabled() || !new.LocalStorage.GetEnabled()},
		{name: "local storage reclaim policy", val: &config.LocalStorage.ReclaimPolicy, old: existing.LocalStorage.ReclaimPolicy, new: new.LocalStorage.ReclaimPolicy, allowChange: !existing.LocalStorage.GetEnabled() || !new.LocalStorage.GetEnabled()},
	} {
		if *i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update string slice fields
	for _, i := range []struct {
		name        string
		val         **[]string
		old         *[]string
		new         *[]string
		allowChange bool
	}{
		{name: "DNS upstream nameservers", val: &config.DNS.UpstreamNameservers, old: existing.DNS.UpstreamNameservers, new: new.DNS.UpstreamNameservers, allowChange: true},
		{name: "load balancer CIDRs", val: &config.LoadBalancer.CIDRs, old: existing.LoadBalancer.CIDRs, new: new.LoadBalancer.CIDRs, allowChange: true},
		{name: "load balancer L2 interfaces", val: &config.LoadBalancer.L2Interfaces, old: existing.LoadBalancer.L2Interfaces, new: new.LoadBalancer.L2Interfaces, allowChange: true},
	} {
		if *i.val, err = mergeSliceField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update int fields
	for _, i := range []struct {
		name        string
		val         **int
		old         *int
		new         *int
		allowChange bool
	}{
		// apiserver
		{name: "kube-apiserver secure port", val: &config.APIServer.SecurePort, old: existing.APIServer.SecurePort, new: new.APIServer.SecurePort},
		// datastore
		{name: "k8s-dqlite port", val: &config.Datastore.K8sDqlitePort, old: existing.Datastore.K8sDqlitePort, new: new.Datastore.K8sDqlitePort},
		// load-balancer
		{name: "load balancer BGP local ASN", val: &config.LoadBalancer.BGPLocalASN, old: existing.LoadBalancer.BGPLocalASN, new: new.LoadBalancer.BGPLocalASN, allowChange: true},
		{name: "load balancer BGP peer ASN", val: &config.LoadBalancer.BGPPeerASN, old: existing.LoadBalancer.BGPPeerASN, new: new.LoadBalancer.BGPPeerASN, allowChange: true},
		{name: "load balancer BGP peer port", val: &config.LoadBalancer.BGPPeerPort, old: existing.LoadBalancer.BGPPeerPort, new: new.LoadBalancer.BGPPeerPort, allowChange: true},
	} {
		if *i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// update bool fields
	for _, i := range []struct {
		name        string
		val         **bool
		old         *bool
		new         *bool
		allowChange bool
	}{
		// network
		{name: "network enabled", val: &config.Network.Enabled, old: existing.Network.Enabled, new: new.Network.Enabled, allowChange: true},
		// DNS
		{name: "DNS enabled", val: &config.DNS.Enabled, old: existing.DNS.Enabled, new: new.DNS.Enabled, allowChange: true},
		// gateway
		{name: "gateway enabled", val: &config.Gateway.Enabled, old: existing.Gateway.Enabled, new: new.Gateway.Enabled, allowChange: true},
		// ingress
		{name: "ingress enabled", val: &config.Ingress.Enabled, old: existing.Ingress.Enabled, new: new.Ingress.Enabled, allowChange: true},
		{name: "ingress enable proxy protocol", val: &config.Ingress.EnableProxyProtocol, old: existing.Ingress.EnableProxyProtocol, new: new.Ingress.EnableProxyProtocol, allowChange: true},
		// load-balancer
		{name: "load balancer enabled", val: &config.LoadBalancer.Enabled, old: existing.LoadBalancer.Enabled, new: new.LoadBalancer.Enabled, allowChange: true},
		{name: "load balancer L2 mode", val: &config.LoadBalancer.L2Mode, old: existing.LoadBalancer.L2Mode, new: new.LoadBalancer.L2Mode, allowChange: true},
		{name: "load balancer BGP mode", val: &config.LoadBalancer.BGPMode, old: existing.LoadBalancer.BGPMode, new: new.LoadBalancer.BGPMode, allowChange: true},
		// local-storage
		{name: "local storage enabled", val: &config.LocalStorage.Enabled, old: existing.LocalStorage.Enabled, new: new.LocalStorage.Enabled, allowChange: true},
		{name: "local storage set default", val: &config.LocalStorage.SetDefault, old: existing.LocalStorage.SetDefault, new: new.LocalStorage.SetDefault, allowChange: true},
		// metrics-server
		{name: "metrics server enabled", val: &config.MetricsServer.Enabled, old: existing.MetricsServer.Enabled, new: new.MetricsServer.Enabled, allowChange: true},
	} {
		if *i.val, err = mergeField(i.old, i.new, i.allowChange); err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	// post check: ensure network is enabled if any of ingress, gateway, load-balancer are enabled
	if !config.Network.GetEnabled() {
		if config.Gateway.GetEnabled() {
			return ClusterConfig{}, fmt.Errorf("gateway requires network to be enabled")
		}
		if config.LoadBalancer.GetEnabled() {
			return ClusterConfig{}, fmt.Errorf("load-balancer requires network to be enabled")
		}
		if config.Ingress.GetEnabled() {
			return ClusterConfig{}, fmt.Errorf("ingress requires network to be enabled")
		}
	}

	// post check: load-balancer BGP mode configuration
	if config.LoadBalancer.GetBGPMode() {
		if config.LoadBalancer.GetBGPLocalASN() == 0 {
			return ClusterConfig{}, fmt.Errorf("load-balancer.bgp-local-asn must be set when load-balancer.bgp-mode is enabled")
		}
		if config.LoadBalancer.GetBGPPeerAddress() == "" {
			return ClusterConfig{}, fmt.Errorf("load-balancer.bgp-peer-address must be set when load-balancer.bgp-mode is enabled")
		}
		if config.LoadBalancer.GetBGPPeerPort() == 0 {
			return ClusterConfig{}, fmt.Errorf("load-balancer.bgp-peer-port must be set when load-balancer.bgp-mode is enabled")
		}
		if config.LoadBalancer.GetBGPPeerASN() == 0 {
			return ClusterConfig{}, fmt.Errorf("load-balancer.bgp-peer-asn must be set when load-balancer.bgp-mode is enabled")
		}
	}

	// post check: local-storage.reclaim-policy should be one of 3 values
	switch config.LocalStorage.GetReclaimPolicy() {
	case "", "Retain", "Recycle", "Delete":
	default:
		return ClusterConfig{}, fmt.Errorf("local-storage.reclaim-policy must be one of: Retrain, Recycle, Delete")
	}

	// post check: local-storage.local-path must be set if enabled
	if config.LocalStorage.GetEnabled() && config.LocalStorage.GetLocalPath() == "" {
		return ClusterConfig{}, fmt.Errorf("local-storage.local-path must be set when local-storage is enabled")
	}

	// post check: ensure cluster DNS is a valid IP address
	if v := config.Kubelet.GetClusterDNS(); v != "" {
		if net.ParseIP(v) == nil {
			return ClusterConfig{}, fmt.Errorf("dns.service-ip must be a valid IP address")
		}

		// TODO: ensure dns.service-ip is part of new.Network.ServiceCIDR
	}

	return config, nil
}
