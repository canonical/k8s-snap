---
myst:
  html_meta:
    description: "Reference documentation for Canonical Kubernetes cluster annotations."
---

# Annotations

This page outlines the annotations that can be configured during cluster
[bootstrap] or when setting cluster configuration later on.

To specify annotations during the bootstrap process, set the
`cluster-config.annotations` parameter in the bootstrap configuration:

```yaml
cluster-config:
...
    annotations:
        k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove: true
        k8sd/v1alpha/lifecycle/skip-stop-services-on-remove: true
```

You can also set annotations after the bootstrap. For example to change
Cilium's VXLAN port you can run the following command:

```bash
sudo k8s set annotations="k8sd/v1alpha1/cilium/tunnel-port=<PORT-NUMBER>"
```

```{note}
v1alpha annotations are experimental and subject to change or removal in future {{product}} releases
```

## `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove`

|                 |                                                                                                                                                                                                                                                   |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | "true"\|"false"                                                                                                                                                                                                                                   |
| **Description** | If set, only Microcluster and file cleanup are performed.  This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default,  k8sd will remove the Kubernetes node when it is removed from the cluster. |

## `k8sd/v1alpha/lifecycle/skip-stop-services-on-remove`

|                 |                                                                                                                                                                                                                                                 |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | "true"\|"false"                                                                                                                                                                                                                                 |
| **Description** | If set, the k8s services will not be stopped on the leaving node when removing the node. This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default, all services are stopped on leaving nodes. |

## `k8sd/v1alpha/lifecycle/disable-separate-feature-upgrades`

|            |                 |
|------------|-----------------|
| **Values** | "true"\|"false" |
|**Description**|If set, the separate feature upgrade is disabled. This is useful, if an external controller (e.g. CAPI) is responsible for the Kubernetes node life cycle. By default, the feature upgrade will be done after all nodes in a cluster are upgraded.|

## `k8sd/v1alpha1/csrsigning/auto-approve`

|                 |                                                                                 |
|-----------------|---------------------------------------------------------------------------------|
| **Values**      | "true"\|"false"                                                                 |
| **Description** | If set, certificate signing requests created by worker nodes are auto approved. |

## `k8sd/v1alpha1/cilium/cni-exclusive`

|                 |                                                                                                                                                                                                                                                                                                         |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | "true"\|"false"                                                                                                                                                                                                                                                                                         |
| **Description** | Make Cilium take ownership over the `/etc/cni/net.d` directory on the node, renaming all non-Cilium CNI configurations to `*.cilium_bak`. This ensures no Pods can be scheduled using other CNI plugins during Cilium agent downtime. Set this to "false" if you wish to use other CNIs such as Multus. |

## `k8sd/v1alpha1/cilium/devices`

|                 |                                                                                                                                                                        |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | string                                                                                                                                                                 |
| **Description** | List of devices facing cluster/external network (used for BPF NodePort, BPF masquerading and host firewall); supports `+` as wildcard in device name, e.g. `eth+,ens+` |

## `k8sd/v1alpha1/cilium/direct-routing-device`

|                 |                                                                                                                                                                                                                                                           |
|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | string                                                                                                                                                                                                                                                    |
| **Description** | Device name used to connect nodes in direct routing mode (used by BPF NodePort, BPF host routing); if empty, automatically set to a device with k8s InternalIP/ExternalIP or with a default route. Bridge type devices are ignored in automatic selection |

## `k8sd/v1alpha1/cilium/sctp/enabled`

|                 |                                 |
|-----------------|---------------------------------|
| **Values**      | "true"\|"false"                 |
| **Description** | Enable the Cilium SCTP feature. |

## `k8sd/v1alpha1/cilium/vlan-bpf-bypass`

|                 |                                                                                                                                                                                                                                                                                                                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | \[] (string values comma separated)                                                                                                                                                                                                                                                                                                                                                |
| **Description** | Comma separated list of VLAN tags to bypass eBPF filtering on native devices. Cilium enables a firewall on native devices and filters all unknown traffic, including VLAN 802.1q packets, which pass through the main device with the associated tag (e.g., VLAN device eth0.4000 and its main interface eth0). Supports `0` as wildcard for bypassing all VLANs. e.g. `4001,4002` |

## `k8sd/v1alpha1/cilium/tunnel-port`

|   |   |
|---|---|
|**Values**| integer value port number|
|**Description**|The port number cilium will for its VXLAN encapsulation protocol destination port.|

## `k8sd/v1alpha1/metrics-server/image-repo`

|                 |                                                               |
|-----------------|---------------------------------------------------------------|
| **Values**      | string                                                        |
| **Description** | Override the default image repository for the metrics-server. |

## `k8sd/v1alpha1/metrics-server/image-tag`

|                 |                                                        |
|-----------------|--------------------------------------------------------|
| **Values**      | string                                                 |
| **Description** | Override the default image tag for the metrics-server. |

## `k8sd/v1alpha1/metallb/bgp-peers`

|                 |                                                                                                                                                                                                                                                                          |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | YAML-encoded list of BGP peer configurations                                                                                                                                                                                                                             |
| **Description** | Configure multi-peer MetalLB BGP mode. When set, this annotation replaces the single-peer typed BGP keys entirely. See [multi-peer BGP]. |

Supported fields per peer:

| Field | Required | Description |
|---|---|---|
| `peerAddress` | yes | IP address of the BGP peer router |
| `peerASN` | yes | ASN of the peer router (1–4294967295) |
| `myASN` | no | Local ASN for this peer (defaults to `load-balancer.bgp-local-asn`) |
| `peerPort` | no | TCP port (default: 179) |
| `nodeSelector` | no | `matchLabels` selector; omit to select all nodes |
| `bfdProfile` | no | Name of a `BFDProfile` to attach to this peer; requires `k8sd/v1alpha1/metallb/bgp-backend` set to `frr-k8s` |

## `k8sd/v1alpha1/metallb/advertise-all-pools`

|                 |                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | "true"\|"false"                                                                                                              |
| **Description** | If set to "true", the BGPAdvertisement is created with an empty spec, which advertises all IP pools. If "false" (default), only the IPAddressPool created by {{product}} is advertised. |

## `k8sd/v1alpha1/metallb/bgp-backend`

|                 |                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | "native"\|"frr-k8s"                                                                                                          |
| **Description** | BGP implementation used by MetalLB. `native` (default) is MetalLB's built-in implementation and does not support BFD. `frr-k8s` runs FRR in a separate `frr-k8s` DaemonSet and is required for BFD. See [multi-peer BGP]. |

## `k8sd/v1alpha1/metallb/bfd-profiles`

|                 |                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------|
| **Values**      | YAML-encoded list of BFD profile configurations                                                                              |
| **Description** | `BFDProfile` resources that {{product}} keeps in sync in the `metallb-system` namespace while BGP mode is enabled. Peers reference them by name in `bfdProfile`. Requires `k8sd/v1alpha1/metallb/bgp-backend` set to `frr-k8s`. Unknown fields are rejected. See [multi-peer BGP]. |

Supported fields per profile (omitted fields use MetalLB's defaults):

| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique profile name (lowercase DNS subdomain) |
| `receiveInterval` | no | Minimum receive interval in ms (10–60000, default: 300) |
| `transmitInterval` | no | Minimum transmit interval in ms (10–60000, default: 300) |
| `detectMultiplier` | no | Missed packets before the session is declared down (2–255, default: 3) |
| `echoInterval` | no | Minimum echo receive interval in ms (10–60000, default: 50) |
| `echoMode` | no | Enable echo mode (default: false) |
| `passiveMode` | no | Wait for the peer to start the session (default: false) |

<script>
const el = document.getElementsByTagName("h2");
for(var i=0;i<el.length;i++){
  el[i].style.fontSize = '1.5em';
  el[i].style.fontWeight = '600';
}
</script>

<!-- Links -->

[bootstrap]: /snap/reference/config-files/bootstrap-config.md
[multi-peer BGP]: /snap/howto/networking/multi-peer-bgp.md
