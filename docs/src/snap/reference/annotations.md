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

Please refer to the [Kubernetes website] for more information on annnotations.

## `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove`

|   |   |
|---|---|
| **Values**| "true"\|"false"|
| **Description**| If set, only MicroCluster and file cleanup are performed.  This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default,  k8sd will remove the Kubernetes node when it is removed from the cluster. |

## `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|If set, the k8s services will not be stopped on the leaving node when removing the node. This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default, all services are stopped on leaving nodes.|

## `k8sd/v1alpha1/csrsigning/auto-approve`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|If set, certificate signing requests created by worker nodes are auto approved.|

## `k8sd/v1alpha1/calico/apiserver-enabled`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|Enable the installation of the Calico API server to enable management of Calico APIs using kubectl.|

## `k8sd/v1alpha1/calico/encapsulation-v4`

|   |   |
|---|---|
|**Values**| “IPIP”\|”VXLAN”\|”IPIPCrossSubnet”\|”VXLANCrossSubnet”\|”None”|
|**Description**|The type of encapsulation to use on the IPv4 pool.|

## `k8sd/v1alpha1/calico/encapsulation-v6`

|   |   |
|---|---|
|**Values**| “IPIP”\|”VXLAN”\|”IPIPCrossSubnet”\|”VXLANCrossSubnet”\|”None”|
|**Description**|The type of encapsulation to use on the IPv6 pool.|

## `k8sd/v1alpha1/calico/autodetection-v4/firstFound`

|   |   |
|---|---|
|**Values**| "true"\|"false"|
|**Description**|Use default interface matching parameters to select an interface, performing best-effort filtering based on well-known interface names.|

## `k8sd/v1alpha1/calico/autodetection-v4/kubernetes`

|   |   |
|---|---|
|**Values**| “NodeInternalIP”|
|**Description**|Configure Calico to detect node addresses based on the Kubernetes API.|

## `k8sd/v1alpha1/calico/autodetection-v4/interface`

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on interfaces that match the given regex.|

## `k8sd/v1alpha1/calico/autodetection-v4/skipInterface`

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on interfaces that do not match the given regex.|

## `k8sd/v1alpha1/calico/autodetection-v4/canReach`

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on which source address on the node is used to reach the specified IP or domain.|

## `k8sd/v1alpha1/calico/autodetection-v4/cidrs`

|   |   |
|---|---|
|**Values**| \[] (string values comma separated) |
|**Description**|Enable IP auto-detection based on which addresses on the nodes are within one of the provided CIDRs.|

## `k8sd/v1alpha1/calico/autodetection-v6/firstFound` 

|   |   |
|---|---|
|**Values**| "true"\|"false" |
|**Description**|Use default interface matching parameters to select an interface, performing best-effort filtering based on well-known interface names.|

## `k8sd/v1alpha1/calico/autodetection-v6/kubernetes`

|   |   |
|---|---|
|**Values**| “NodeInternalIP” |
|**Description**|Configure Calico to detect node addresses based on the Kubernetes API.|

## `k8sd/v1alpha1/calico/autodetection-v6/interface`

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on interfaces that match the given regex.|

## `k8sd/v1alpha1/calico/autodetection-v6/skipInterface`

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on interfaces that do not match the given regex.|

## `k8sd/v1alpha1/calico/autodetection-v6/canReach` 

|   |   |
|---|---|
|**Values**| string |
|**Description**|Enable IP auto-detection based on which source address on the node is used to reach the specified IP or domain.|

## `k8sd/v1alpha1/calico/autodetection-v6/cidrs`

|   |   |
|---|---|
|**Values**| \[] (string values comma separated) |
|**Description**|Enable IP auto-detection based on which addresses on the nodes are within one of the provided CIDRs.|

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
