# Annotations

This page outlines the annotations that can be configured during cluster
[bootstrap]. To do this, set the cluster-config/annotations parameter in
the bootstrap configuration.

| Name                                                          | Description                                                                                                                                                                                                                                       | Values          |
|---------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove` | If set, only microcluster and file cleanup are performed.  This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default,  k8sd will remove the Kubernetes node when it is removed from the cluster. | "true"\|"false" |
| `k8sd/v1alpha/lifecycle/skip-stop-services-on-remove` | If set, the k8s services will not be stopped on the leaving node when removing the node. This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default, all services are stopped on leaving nodes. | "true"\|"false" |
| `k8sd/v1alpha1/csrsigning/auto-approve` | If set, certificate signing requests created by worker nodes are auto approved. | "true"\|"false" |
| `k8sd/v1alpha1/calico/apiserver-enabled` | Enable the installation of the Calico API server to enable management of Calico APIs using kubectl. | "true"\|"false" |
| `k8sd/v1alpha1/calico/encapsulation-v4` | The type of encapsulation to use on the IPv4 pool. | "IPIP"\|"VXLAN"\|"IPIPCrossSubnet"\|"VXLANCrossSubnet"\|"None" |
| `k8sd/v1alpha1/calico/encapsulation-v6` | The type of encapsulation to use on the IPv6 pool. | "IPIP"\|"VXLAN"\|"IPIPCrossSubnet"\|"VXLANCrossSubnet"\|"None" |
| `k8sd/v1alpha1/calico/autodetection-v4/firstFound` | Use default interface matching parameters to select an interface, performing best-effort filtering based on well-known interface names. | "true"\|"false" |
| `k8sd/v1alpha1/calico/autodetection-v4/kubernetes` | Configure Calico to detect node addresses based on the Kubernetes API. | "NodeInternalIP" |
| `k8sd/v1alpha1/calico/autodetection-v4/interface` | Enable IP auto-detection based on interfaces that match the given regex. | string |
| `k8sd/v1alpha1/calico/autodetection-v4/skipInterface` | Enable IP auto-detection based on interfaces that do not match the given regex. | string |
| `k8sd/v1alpha1/calico/autodetection-v4/canReach` | Enable IP auto-detection based on which source address on the node is used to reach the specified IP or domain. | string |
| `k8sd/v1alpha1/calico/autodetection-v4/cidrs` | Enable IP auto-detection based on which addresses on the nodes are within one of the provided CIDRs. | []string (comma separated) |
| `k8sd/v1alpha1/calico/autodetection-v6/firstFound` | Use default interface matching parameters to select an interface, performing best-effort filtering based on well-known interface names. | "true"\|"false" |
| `k8sd/v1alpha1/calico/autodetection-v6/kubernetes` | Configure Calico to detect node addresses based on the Kubernetes API. | "NodeInternalIP" |
| `k8sd/v1alpha1/calico/autodetection-v6/interface` | Enable IP auto-detection based on interfaces that match the given regex. | string |
| `k8sd/v1alpha1/calico/autodetection-v6/skipInterface` | Enable IP auto-detection based on interfaces that do not match the given regex. | string |
| `k8sd/v1alpha1/calico/autodetection-v6/canReach` | Enable IP auto-detection based on which source address on the node is used to reach the specified IP or domain. | string |
| `k8sd/v1alpha1/calico/autodetection-v6/cidrs` | Enable IP auto-detection based on which addresses on the nodes are within one of the provided CIDRs. | []string (comma separated) |
| `k8sd/v1alpha1/cilium/devices` | List of devices facing cluster/external network (used for BPF NodePort, BPF masquerading and host firewall); supports `+` as wildcard in device name, e.g. `eth+,ens+` | string |
| `k8sd/v1alpha1/cilium/direct-routing-device` | Device name used to connect nodes in direct routing mode (used by BPF NodePort, BPF host routing); if empty, automatically set to a device with k8s InternalIP/ExternalIP or with a default route. Bridge type devices are ignored in automatic selection | string |
| `k8sd/v1alpha1/cilium/vlan-bpf-bypass` | Comma separated list of VLAN tags to bypass eBPF filtering on native devices. Cilium enables firewalling on native devices and filters all unknown traffic, including VLAN 802.1q packets, which pass through the main device with the associated tag (e.g., VLAN device eth0.4000 and its main interface eth0). Supports `0` as wildcard for bypassing all VLANs. e.g. `4001,4002` | []string |
| `k8sd/v1alpha1/metrics-server/image-repo` | Override the default image repository for the metrics-server. | string |
| `k8sd/v1alpha1/metrics-server/image-tag` | Override the default image tag for the metrics-server. | string |


<!-- Links -->

[bootstrap]: /snap/reference/bootstrap-config-reference
