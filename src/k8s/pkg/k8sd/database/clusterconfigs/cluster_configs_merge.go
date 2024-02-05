package clusterconfigs

import "fmt"

func mergeValue[T comparable](old T, new T, allowChange bool) (T, error) {
	var zeroValue T
	if old != zeroValue && new != zeroValue && new != old && !allowChange {
		return zeroValue, fmt.Errorf("value has changed")
	}
	if old == zeroValue {
		return new, nil
	}
	return old, nil
}

// Merge applies updates from non-empty values of the new ClusterConfig to an existing one.
// Merge will return an error if we try to update a config that must not be updated. once such an operation is implemented in the future, we can allow the change here.
// Merge will create a new ClusterConfig object to avoid mutating the existing config objects.
func Merge(existing ClusterConfig, new ClusterConfig) (ClusterConfig, error) {
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
		{name: "apiserver-to-kubelet certificate", val: &config.Certificates.APIServerToKubeletCert, old: existing.Certificates.APIServerToKubeletCert, new: new.Certificates.APIServerToKubeletCert, allowChange: true},
		{name: "apiserver-to-kubelet key", val: &config.Certificates.APIServerToKubeletKey, old: existing.Certificates.APIServerToKubeletKey, new: new.Certificates.APIServerToKubeletKey, allowChange: true},
		{name: "front proxy CA certificate", val: &config.Certificates.FrontProxyCACert, old: existing.Certificates.FrontProxyCACert, new: new.Certificates.FrontProxyCACert, allowChange: true},
		{name: "front proxy CA key", val: &config.Certificates.FrontProxyCAKey, old: existing.Certificates.FrontProxyCAKey, new: new.Certificates.FrontProxyCAKey, allowChange: true},
		{name: "authorization-mode", val: &config.APIServer.AuthorizationMode, old: existing.APIServer.AuthorizationMode, new: new.APIServer.AuthorizationMode, allowChange: true},
		{name: "service account key", val: &config.APIServer.ServiceAccountKey, old: existing.APIServer.ServiceAccountKey, new: new.APIServer.ServiceAccountKey},
		{name: "cluster cidr", val: &config.Cluster.CIDR, old: existing.Cluster.CIDR, new: new.Cluster.CIDR},
		{name: "datastore", val: &config.APIServer.Datastore, old: existing.APIServer.Datastore, new: new.APIServer.Datastore, allowChange: true},
		{name: "datastore url", val: &config.APIServer.DatastoreURL, old: existing.APIServer.DatastoreURL, new: new.APIServer.DatastoreURL, allowChange: true},
		{name: "datastore ca", val: &config.APIServer.DatastoreCA, old: existing.APIServer.DatastoreCA, new: new.APIServer.DatastoreCA, allowChange: true},
		{name: "datastore client certificate", val: &config.APIServer.DatastoreClientCert, old: existing.APIServer.DatastoreClientCert, new: new.APIServer.DatastoreClientCert, allowChange: true},
		{name: "datastore client key", val: &config.APIServer.DatastoreClientKey, old: existing.APIServer.DatastoreClientKey, new: new.APIServer.DatastoreClientKey, allowChange: true},
		{name: "cluster dns", val: &config.Kubelet.ClusterDNS, old: existing.Kubelet.ClusterDNS, new: new.Kubelet.ClusterDNS, allowChange: true},
		{name: "cluster domain", val: &config.Kubelet.ClusterDomain, old: existing.Kubelet.ClusterDomain, new: new.Kubelet.ClusterDomain, allowChange: true},
		{name: "cloud provider", val: &config.Kubelet.CloudProvider, old: existing.Kubelet.CloudProvider, new: new.Kubelet.CloudProvider, allowChange: true},
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
	} {
		*i.val, err = mergeValue(i.old, i.new, i.allowChange)
		if err != nil {
			return ClusterConfig{}, fmt.Errorf("prevented update of %s: %w", i.name, err)
		}
	}

	return config, nil
}
