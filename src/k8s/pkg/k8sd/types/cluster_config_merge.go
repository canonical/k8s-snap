package types

import (
	"fmt"
	"slices"
)

func mergeValue[T comparable](old T, new T, allowChange bool) (T, error) {
	var zeroValue T
	if old != zeroValue && new != zeroValue && new != old && !allowChange {
		return zeroValue, fmt.Errorf("value has changed")
	}
	if new != zeroValue {
		return new, nil
	}
	return old, nil
}

func mergeSlice[T comparable](old []T, new []T, allowChange bool) ([]T, error) {
	if old != nil && new != nil && slices.Equal(old, new) && !allowChange {
		return nil, fmt.Errorf("value has changed")
	}
	if new != nil {
		return new, nil
	}
	return old, nil
}

// MergeClusterConfig applies updates from non-empty values of the new ClusterConfig to an existing one.
// MergeClusterConfig will return an error if we try to update a config that must not be updated. once such an operation is implemented in the future, we can allow the change here.
// MergeClusterConfig will create a new ClusterConfig object to avoid mutating the existing config objects.
func MergeClusterConfig(existing ClusterConfig, new ClusterConfig) (ClusterConfig, error) {
	var (
		config ClusterConfig
		err    error
	)

	for _, i := range []struct {
		name        string
		val         *string
		old         string
		new         string
		allowChange bool
	}{
		{name: "cluster CA certificate", val: &config.Certificates.CACert, old: existing.Certificates.CACert, new: new.Certificates.CACert},
		{name: "cluster CA key", val: &config.Certificates.CAKey, old: existing.Certificates.CAKey, new: new.Certificates.CAKey},
		{name: "k8s-dqlite certificate", val: &config.Certificates.K8sDqliteCert, old: existing.Certificates.K8sDqliteCert, new: new.Certificates.K8sDqliteCert},
		{name: "k8s-dqlite key", val: &config.Certificates.K8sDqliteKey, old: existing.Certificates.K8sDqliteKey, new: new.Certificates.K8sDqliteKey},
		{name: "apiserver-kubelet-client certificate", val: &config.Certificates.APIServerKubeletClientCert, old: existing.Certificates.APIServerKubeletClientCert, new: new.Certificates.APIServerKubeletClientCert, allowChange: true},
		{name: "apiserver-kubelet-client key", val: &config.Certificates.APIServerKubeletClientKey, old: existing.Certificates.APIServerKubeletClientKey, new: new.Certificates.APIServerKubeletClientKey, allowChange: true},
		{name: "front proxy CA certificate", val: &config.Certificates.FrontProxyCACert, old: existing.Certificates.FrontProxyCACert, new: new.Certificates.FrontProxyCACert, allowChange: true},
		{name: "front proxy CA key", val: &config.Certificates.FrontProxyCAKey, old: existing.Certificates.FrontProxyCAKey, new: new.Certificates.FrontProxyCAKey, allowChange: true},
		{name: "authorization-mode", val: &config.APIServer.AuthorizationMode, old: existing.APIServer.AuthorizationMode, new: new.APIServer.AuthorizationMode, allowChange: true},
		{name: "service account key", val: &config.APIServer.ServiceAccountKey, old: existing.APIServer.ServiceAccountKey, new: new.APIServer.ServiceAccountKey},
		{name: "pod cidr", val: &config.Network.PodCIDR, old: existing.Network.PodCIDR, new: new.Network.PodCIDR},
		{name: "service cidr", val: &config.Network.ServiceCIDR, old: existing.Network.ServiceCIDR, new: new.Network.ServiceCIDR},
		{name: "datastore", val: &config.APIServer.Datastore, old: existing.APIServer.Datastore, new: new.APIServer.Datastore},
		{name: "datastore url", val: &config.APIServer.DatastoreURL, old: existing.APIServer.DatastoreURL, new: new.APIServer.DatastoreURL, allowChange: true},
		{name: "datastore ca", val: &config.APIServer.DatastoreCA, old: existing.APIServer.DatastoreCA, new: new.APIServer.DatastoreCA, allowChange: true},
		{name: "datastore client certificate", val: &config.APIServer.DatastoreClientCert, old: existing.APIServer.DatastoreClientCert, new: new.APIServer.DatastoreClientCert, allowChange: true},
		{name: "datastore client key", val: &config.APIServer.DatastoreClientKey, old: existing.APIServer.DatastoreClientKey, new: new.APIServer.DatastoreClientKey, allowChange: true},
		{name: "cluster dns", val: &config.Kubelet.ClusterDNS, old: existing.Kubelet.ClusterDNS, new: new.Kubelet.ClusterDNS, allowChange: true},
		{name: "cluster domain", val: &config.Kubelet.ClusterDomain, old: existing.Kubelet.ClusterDomain, new: new.Kubelet.ClusterDomain, allowChange: true},
		{name: "cloud provider", val: &config.Kubelet.CloudProvider, old: existing.Kubelet.CloudProvider, new: new.Kubelet.CloudProvider, allowChange: true},

		{name: "ingress.default-tls-secret", val: &config.Ingress.DefaultTLSSecret, old: existing.Ingress.DefaultTLSSecret, new: new.Ingress.DefaultTLSSecret, allowChange: true},

		{name: "load-balancer.bgp-peer-address", val: &config.LoadBalancer.BGPPeerAddress, old: existing.LoadBalancer.BGPPeerAddress, new: new.LoadBalancer.BGPPeerAddress, allowChange: true},

		{name: "local-storage.local-path", val: &config.LocalStorage.LocalPath, old: existing.LocalStorage.LocalPath, new: new.LocalStorage.LocalPath, allowChange: true},
		{name: "local-storage.set-default", val: &config.LocalStorage.ReclaimPolicy, old: existing.LocalStorage.ReclaimPolicy, new: new.LocalStorage.ReclaimPolicy, allowChange: true},
	} {
		*i.val, err = mergeValue(i.old, i.new, i.allowChange)
		if err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	for _, i := range []struct {
		name        string
		val         *int
		old         int
		new         int
		allowChange bool
	}{
		{name: "secure port", val: &config.APIServer.SecurePort, old: existing.APIServer.SecurePort, new: new.APIServer.SecurePort},
		{name: "k8s-dqlite port", val: &config.K8sDqlite.Port, old: existing.K8sDqlite.Port, new: new.K8sDqlite.Port},

		{name: "load-balancer.bgp-local-asn", val: &config.LoadBalancer.BGPLocalASN, old: existing.LoadBalancer.BGPLocalASN, new: new.LoadBalancer.BGPLocalASN, allowChange: true},
		{name: "load-balancer.bgp-peer-asn", val: &config.LoadBalancer.BGPPeerASN, old: existing.LoadBalancer.BGPPeerASN, new: new.LoadBalancer.BGPPeerASN, allowChange: true},
		{name: "load-balancer.bgp-peer-port", val: &config.LoadBalancer.BGPPeerPort, old: existing.LoadBalancer.BGPPeerPort, new: new.LoadBalancer.BGPPeerPort, allowChange: true},
	} {
		*i.val, err = mergeValue(i.old, i.new, i.allowChange)
		if err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	for _, i := range []struct {
		name        string
		val         *[]string
		old         []string
		new         []string
		allowChange bool
	}{
		{name: "dns.upstream-nameservers", val: &config.DNS.UpstreamNameservers, old: existing.DNS.UpstreamNameservers, new: new.DNS.UpstreamNameservers, allowChange: true},

		{name: "load-balancer.cidrs", val: &config.LoadBalancer.CIDRs, old: existing.LoadBalancer.CIDRs, new: new.LoadBalancer.CIDRs, allowChange: true},
		{name: "load-balancer.l2-interfaces", val: &config.LoadBalancer.L2Interfaces, old: existing.LoadBalancer.L2Interfaces, new: new.LoadBalancer.L2Interfaces, allowChange: true},
	} {
		*i.val, err = mergeSlice(i.old, i.new, i.allowChange)
		if err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	for _, i := range []struct {
		name        string
		val         **bool
		old         *bool
		new         *bool
		allowChange bool
	}{
		{name: "network.enabled", val: &config.Network.Enabled, old: existing.Network.Enabled, new: new.Network.Enabled, allowChange: true},
		{name: "dns.enabled", val: &config.DNS.Enabled, old: existing.DNS.Enabled, new: new.DNS.Enabled, allowChange: true},
		{name: "gateway.enabled", val: &config.Gateway.Enabled, old: existing.Gateway.Enabled, new: new.Gateway.Enabled, allowChange: true},
		{name: "ingress.enabled", val: &config.Ingress.Enabled, old: existing.Ingress.Enabled, new: new.Ingress.Enabled, allowChange: true},
		{name: "load-balancer.enabled", val: &config.LoadBalancer.Enabled, old: existing.LoadBalancer.Enabled, new: new.LoadBalancer.Enabled, allowChange: true},
		{name: "local-storage.enabled", val: &config.LocalStorage.Enabled, old: existing.LocalStorage.Enabled, new: new.LocalStorage.Enabled, allowChange: true},
		{name: "metrics-server.enabled", val: &config.MetricsServer.Enabled, old: existing.MetricsServer.Enabled, new: new.MetricsServer.Enabled, allowChange: true},

		{name: "ingress.enable-proxy-protocol", val: &config.Ingress.EnableProxyProtocol, old: existing.Ingress.EnableProxyProtocol, new: new.Ingress.EnableProxyProtocol, allowChange: true},

		{name: "load-balancer.l2-mode", val: &config.LoadBalancer.L2Enabled, old: existing.LoadBalancer.L2Enabled, new: new.LoadBalancer.L2Enabled, allowChange: true},
		{name: "load-balancer.bgp-mode", val: &config.LoadBalancer.BGPEnabled, old: existing.LoadBalancer.BGPEnabled, new: new.LoadBalancer.BGPEnabled, allowChange: true},

		{name: "local-storage.set-default", val: &config.LocalStorage.SetDefault, old: existing.LocalStorage.SetDefault, new: new.LocalStorage.SetDefault, allowChange: true},
	} {
		*i.val, err = mergeValue(i.old, i.new, i.allowChange)
		if err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	return config, nil
}
