# How to use an alternative CNI

While {{product}} ships with a default [Container Network Interface] (CNI) that
we ensure is fully compatible with our distribution, it's possible to use a
different CNI plugin for your specific networking requirements. This guide
explains how to safely replace the default CNI with an alternative solution.

## Prerequisites

This guide assumes the following:

- Root or sudo access to the machine.
- Basic understanding of Kubernetes networking concepts.
- Basic knowledge of [Helm].

## Disable default network implementation

For an existing cluster, disable the default network
plugin:

```
sudo k8s disable ingress gateway network
```

For a new cluster, create a bootstrap configuration that disables networking:

```
cat <<EOF > bootstrap-config.yaml
cluster-config:
  network:
    enabled: false
```

Then, bootstrap the cluster with this configuration:

```
sudo k8s bootstrap --file bootstrap-config.yaml
```

## Configure Helm repository

Add the CNI's Helm repository to {{product}}'s Helm installation. This guide
uses [Calico] as an example:

```
sudo k8s helm repo add projectcalico https://docs.tigera.io/calico/charts
```

## Install Alternative CNI

Create a values file with the basic configuration for Calico:

```
cat <<EOF > values.yaml
apiServer:
  enabled: false
calicoctl:
  image: ghcr.io/canonical/k8s-snap/calico/ctl
  tag: v3.28.0
installation:
  calicoNetwork:
    ipPools:
    - cidr: 10.1.0.0/16
      encapsulation: VXLAN
      name: ipv4-ippool
  registry: ghcr.io/canonical/k8s-snap
serviceCIDRs:
- 10.152.183.0/24
tigeraOperator:
  image: tigera/operator
  registry: ghcr.io/canonical/k8s-snap
  version: v1.34.0
EOF
```

After saving the values file, create the required namespace:

```
sudo k8s kubectl create namespace tigera-operator
```

Deploy Calico using Helm:

```
sudo k8s helm install calico projectcalico/tigera-operator --version v3.28.0 -f values.yaml --namespace tigera-operator
```

## Verify deployment

Monitor the status of the calico pods:

```
watch sudo k8s kubectl get pods -n calico-system
```

If Calico is deployed successfully, the output will be similar to:

```
NAME                                       READY   STATUS    RESTARTS   AGE
calico-kube-controllers-7bc846689c-9p2kp   1/1     Running   0          22h
calico-node-2bm8m                          1/1     Running   0          22h
calico-typha-56f55cb75-cj2jk               1/1     Running   0          22h
csi-node-driver-vth9t                      2/2     Running   0          22h
```

## Reverting

If the deployment does not work as expected, you can always revert to the
default networking configuration.

Remove all resources associated with Calico:

```
sudo k8s helm uninstall calico --namespace tigera-operator
```

Remove the alternative CNI's namespace:

```
sudo k8s kubectl delete namespace tigera-operator
```

Enable the default networking features:

```
sudo k8s enable ingress gateway network
```

<!-- Links -->
[Container Network Interface]: https://github.com/containernetworking/cni
[Calico]: https://docs.tigera.io/
[Helm]: https://helm.sh/docs
