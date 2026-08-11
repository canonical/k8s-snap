---
myst:
  html_meta:
    description: "How to replace the default CNI in Canonical Kubernetes with an alternative Container Network Interface (CNI) plugin."
---

<!-- SPREAD SUITE: snap_bootstrapped -->

<!-- SPREAD
trap 'sudo rm -f values.yaml' EXIT
-->

# How to use an alternative CNI

{{product}} ships with a default [Container Network Interface](https://github.com/containernetworking/cni) (CNI) that
is fully compatible with our distribution. It is possible, however, to use a
different CNI plugin for your specific networking requirements. This guide
explains how to safely disable the default CNI so you can then deploy your own networking solution and how to revert to the default solution if needed.

## Prerequisites

This guide assumes the following:

- Root or sudo access to the machine.

## Disable default network

### Disable on an existing cluster

For an existing cluster, disable the default network plugin and its 
dependencies:

```{note}
`ingress` and `gateway` must be disabled before or at the same time as `network`.
```

```
sudo k8s disable ingress gateway network
```

<!-- SPREAD
sudo k8s get network | grep "enabled: false"
sudo k8s get ingress | grep "enabled: false"
sudo k8s get gateway | grep "enabled: false"
-->

Remove any configuration files left behind by the default Cilium implementation in `/etc/cni/net.d/`:

```
sudo rm /etc/cni/net.d/05-cilium.conflist
```

### Disable on new clusters

For a new cluster, create a bootstrap configuration that disables networking:

<!-- SPREAD SKIP -->

```
cat <<EOF > bootstrap-config.yaml
cluster-config:
  network:
    enabled: false
EOF
```

Then, bootstrap the cluster with this configuration:

```
sudo k8s bootstrap --file bootstrap-config.yaml
```
<!-- SPREAD SKIP END -->

## Install your CNI

Refer to the instructions provided by your CNI solution for installation. If your CNI
uses `kubectl`, {{product}} supplies a built-in `k8s kubectl` implementation that can help with the installation process.

<!-- SPREAD
sudo k8s helm repo add projectcalico https://docs.tigera.io/calico/charts
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
sudo k8s kubectl create namespace tigera-operator
sudo k8s kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/operator-crds.yaml
sudo k8s helm install calico projectcalico/tigera-operator --version v3.28.0 -f values.yaml --namespace tigera-operator
-->

## Verify deployment

Ensure the alternative CNI pods are running and in a `Ready` state:

<!-- SPREAD SKIP -->

```
watch sudo k8s kubectl get pods -A
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
source ${SPREAD_PATH}/docs/tools/repeat_checks.sh
repeat_checks "sudo k8s kubectl get pods -n calico-system" "calico-kube"
sudo k8s kubectl rollout status deployment/calico-kube-controllers -n calico-system --timeout=300s
sudo k8s kubectl rollout status deployment/calico-typha -n calico-system --timeout=300s
sudo k8s kubectl rollout status daemonset/calico-node -n calico-system --timeout=300s
-->

Once the CNI pods are healthy, verify that cluster nodes reach the `Ready`
state:

```
sudo k8s kubectl get nodes
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

## Revert to default

You can always revert to the default networking configuration. Remove all resources associated with the alternative CNI (pods, namespace, Helm charts etc.). 

<!-- SPREAD
sudo k8s kubectl delete installation default --ignore-not-found
sudo rm -f /etc/cni/net.d/10-calico.conflist
sudo k8s kubectl delete pods --all -n calico-system --force --grace-period=0 2>/dev/null
sudo k8s helm uninstall calico -n tigera-operator 2>/dev/null
sudo k8s kubectl delete ns tigera-operator calico-system --ignore-not-found --timeout=15s
-->

Then enable the default networking features:

```
sudo k8s enable ingress gateway network
```

<!-- SPREAD
sudo k8s get network | grep "enabled: true"
sudo k8s get ingress | grep "enabled: true"
sudo k8s get gateway | grep "enabled: true"
-->

Kubernetes may add a `network-unavailable` taint when no CNI is active. This taint prevents pods from scheduling.

```
sudo k8s kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints
```

Remove the `network-unavailable` taint if present.

<!-- SPREAD SKIP -->

```
sudo k8s kubectl taint nodes --all node.kubernetes.io/network-unavailable-
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
sudo k8s kubectl taint nodes --all node.kubernetes.io/network-unavailable- || true
-->

Watch the `Cilium` pods deploy and reach a `Ready` state.

<!-- SPREAD SKIP -->

```
watch sudo k8s kubectl get pods -n kube-system
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
sudo k8s kubectl rollout status daemonset/cilium -n kube-system --timeout=10m
sudo k8s kubectl taint nodes --all node.kubernetes.io/network-unavailable- || true
sudo k8s kubectl rollout status deployment/cilium-operator -n kube-system --timeout=10m
sudo k8s kubectl taint nodes --all node.kubernetes.io/network-unavailable- || true
sudo k8s kubectl wait --for=condition=Ready pods --all -n kube-system --timeout=10m
-->

