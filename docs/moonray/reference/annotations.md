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
