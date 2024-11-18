# Bootstrap configuration file reference

A YAML file can be supplied to the `k8s join-cluster` command to configure and
customise the cluster. This reference section provides the format of this file
by listing all available options and their details. See below for an example.

## Configuration options

```{include} ../../_parts/bootstrap_config.md
```


## Example

The following example configures and enables certain features, sets an external
cloud provider, marks the control plane nodes as unschedulable, changes the pod
and service CIDRs from the defaults and adds an extra SAN to the generated
certificates.

```yaml
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
    cluster-domain: cluster.local
  ingress:
    enabled: true
  load-balancer:
    enabled: true
    cidrs:
    - 10.0.0.0/24
    - 10.1.0.10-10.1.0.20
    l2-mode: true
  local-storage:
    enabled: true
    local-path: /storage/path
    default: false
  gateway:
    enabled: true
  metrics-server:
    enabled: true
  cloud-provider: external
control-plane-taints:
- node-role.kubernetes.io/control-plane:NoSchedule
pod-cidr: 10.100.0.0/16
service-cidr: 10.200.0.0/16
disable-rbac: false
secure-port: 6443
k8s-dqlite-port: 9090
datastore-type: k8s-dqlite
extra-sans:
- custom.kubernetes
extra-node-config-files:
  bootstrap-extra-file.yaml: extra-args-test-file-content
extra-node-kube-apiserver-args:
  --request-timeout: 2m
extra-node-kube-controller-manager-args:
  --leader-elect-retry-period: 3s
extra-node-kube-scheduler-args:
  --authorization-webhook-cache-authorized-ttl: 11s
extra-node-kube-proxy-args:
  --config-sync-period: 14m
extra-node-kubelet-args:
  --authentication-token-webhook-cache-ttl: 3m
extra-node-containerd-args:
  --log-level: debug
extra-node-k8s-dqlite-args:
  --watch-storage-available-size-interval: 6s
```
