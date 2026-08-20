---
myst:
  html_meta:
    description: "Upgrade notes for Canonical Kubernetes snap, including version-specific migration steps and compatibility changes between releases."
---

# Upgrade notes

## Upgrade 1.35 to 1.36

### Datastore upgrade considerations

{{product}} 1.36 removes the `k8s-dqlite` datastore that was deprecated in
1.35. The `post-refresh` hook rejects the upgrade if the cluster still uses it,
and there is no migration path. Check the datastore in use before refreshing:

```
sudo k8s status
```

If the `datastore` line reports anything other than `etcd`, `external` or
`disabled`, stay on 1.35 or deploy a new cluster backed by etcd.

### Perform the upgrade

```
sudo snap refresh k8s --channel=1.36-classic/stable
```

All components will be updated automatically, including Cilium (1.20),
MetalLB (0.16.1) and the Gateway API CRDs (v1.6.1).

### Verify the upgrade

Check the `k8s` snap version has been updated and the cluster is back in the
`Ready` state.

```
snap info k8s
sudo k8s status --wait-ready
```

### Optional: switch to the Cilium kube-proxy replacement

Clusters bootstrapped on 1.36 with the default Cilium network do not deploy
`kube-proxy`. Instead, Cilium handles service routing in eBPF. Upgraded
clusters keep running `kube-proxy`: the refresh deliberately leaves the
existing datapath in place, so moving to the eBPF replacement is an explicit
opt-in.

To adopt the kube-proxy replacement on an upgraded cluster:

```
sudo k8s set network.kube-proxy-enabled=false
```

Each node stops its `kube-proxy` service and removes the iptables rules it
installed. Verify that services are still reachable and that `kube-proxy` is no
longer running:

```
sudo k8s status
sudo snap services k8s.kube-proxy
```

```{note}
This changes how service traffic is routed on every node. Roll it out during a
maintenance window and validate your workloads afterwards.
```

```{warning}
Disabling `kube-proxy` on 1.36 is not reversible while the network feature is
enabled. If you try to set `network.kube-proxy-enabled` back to `true`, it will
be rejected as long as the default Cilium network is in use.
```

## Upgrade 1.34 to 1.35

```{note}
{{product}} 1.35 upgrades the container runtime to containerd v2. If you
maintain custom containerd configuration or tooling, validate compatibility
before refreshing.
```

Simply run:

```
sudo snap refresh k8s --channel=1.35-classic/stable
```

All components will be updated automatically.

### Verify the upgrade

Check the `k8s` snap version has been updated and the cluster is back in the
`Ready` state.

```
snap info k8s
sudo k8s status --wait-ready
```

## Upgrade 1.33 to 1.34

Simply run:

```bash
sudo snap refresh k8s --channel=1.34-classic/stable
```

All components will be updated automatically.

### Verify the upgrade

Check the `k8s` snap version has been updated and the cluster is back in the
`Ready` state.

```
snap info k8s
sudo k8s status --wait-ready
```

## Upgrade 1.32 to 1.33

If you are not using dual stack networking, you can simply run:

```bash
sudo snap refresh k8s --channel=1.33-classic/stable
```

All components will be updated automatically.

### Additional steps for dual-stack environments

If your cluster is configured with dual stack networking (IPv4 and IPv6),
you’ll need to make a manual adjustment before refreshing. {{product}} 1.33
includes Cilium v1.17, which introduces a stricter requirement for dual stack:
each node must report both IPv4 and IPv6 addresses to the API server.
If this is not satisfied, the Cilium agent pods will fail to start.
For each node in the cluster:

- Update the `--node-ip` flag in the kubelet configuration file
`/var/snap/k8s/common/args/kubelet` to include both the IPv4 and IPv6 addresses
(comma-separated) from the network interface that is used to connect the node
to the cluster network:

```bash
--node-ip=<IPv4>,<IPv6>
```

- Restart the `kubelet` service

```bash
sudo systemctl restart snap.k8s.kubelet.service
```

- Restart the Cilium DaemonSet:

```bash
sudo k8s kubectl rollout restart daemonset cilium -n kube-system
```

Now you can run the snap `refresh` command to perform the upgrade.

### Verify the upgrade

Check the `k8s` snap version has been updated and the cluster is back in the
`Ready` state.

```
snap info k8s
sudo k8s status --wait-ready
```
