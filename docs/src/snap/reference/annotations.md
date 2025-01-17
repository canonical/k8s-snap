# Annotations

This page outlines the annotations that can be configured during cluster
[bootstrap]. To do this, set the `cluster-config.annotations` parameter in
the bootstrap configuration.

For example:

```yaml
cluster-config:
...
    annotations:
        k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove: true
        k8sd/v1alpha/lifecycle/skip-stop-services-on-remove: true
```

```{note}
v1alpha annotations are experimental and subject to change or removal in future {{product}} releases
```

## `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove`

|   |   |
|---|---|
| **Values**| "true"\|"false"|
| **Description**| If set, only MicroCluster and file cleanup are performed.  This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default,  k8sd will remove the Kubernetes node when it is removed from the cluster. |

## `k8sd/v1alpha/lifecycle/skip-stop-services-on-remove`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|If set, the k8s services will not be stopped on the leaving node when removing the node. This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default, all services are stopped on leaving nodes.|

## `k8sd/v1alpha1/csrsigning/auto-approve`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|If set, certificate signing requests created by worker nodes are auto approved.|

## `k8sd/v1alpha1/cilium/cni-exclusive`

|   |   |
|---|---|
| **Values**| "true"\|"false"|
| **Description**| Make Cilium take ownership over the `/etc/cni/net.d` directory on the node, renaming all non-Cilium CNI configurations to `*.cilium_bak`. This ensures no Pods can be scheduled using other CNI plugins during Cilium agent downtime. Set this to "false" if you wish to use other CNIs such as Multus. |

## `k8sd/v1alpha1/cilium/devices`

|   |   |
|---|---|
|**Values**| string|
|**Description**|List of devices facing cluster/external network (used for BPF NodePort, BPF masquerading and host firewall); supports `+` as wildcard in device name, e.g. `eth+,ens+` |

## `k8sd/v1alpha1/cilium/direct-routing-device`

|   |   |
|---|---|
|**Values**| string|
|**Description**|Device name used to connect nodes in direct routing mode (used by BPF NodePort, BPF host routing); if empty, automatically set to a device with k8s InternalIP/ExternalIP or with a default route. Bridge type devices are ignored in automatic selection|

## `k8sd/v1alpha1/cilium/vlan-bpf-bypass`

|   |   |
|---|---|
|**Values**| \[] (string values comma separated)|
|**Description**|Comma separated list of VLAN tags to bypass eBPF filtering on native devices. Cilium enables a firewall on native devices and filters all unknown traffic, including VLAN 802.1q packets, which pass through the main device with the associated tag (e.g., VLAN device eth0.4000 and its main interface eth0). Supports `0` as wildcard for bypassing all VLANs. e.g. `4001,4002`|

## `k8sd/v1alpha1/metrics-server/image-repo`

|   |   |
|---|---|
|**Values**| string|
|**Description**|Override the default image repository for the metrics-server.|

## `k8sd/v1alpha1/metrics-server/image-tag`

|   |   |
|---|---|
|**Values**| string|
|**Description**|Override the default image tag for the metrics-server.|

<script>
const el = document.getElementsByTagName("h2");
for(var i=0;i<el.length;i++){
  el[i].style.fontSize = '1.5em';
  el[i].style.fontWeight = '600';
}
</script>

<!-- Links -->

[Kubernetes website]:https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
[bootstrap]: bootstrap-config-reference
