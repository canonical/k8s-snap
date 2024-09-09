# Annotations

This page outlines the annotations that can be configured during cluster
[bootstrap]. To do this, set the cluster-config/annotations parameter in
the bootstrap configuration.

| Name                                                          | Description                                                                                                                                                                                                                                       | Values          |
|---------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `k8sd/v1alpha/lifecycle/skip-cleanup-kubernetes-node-on-remove` | If set, only microcluster and file cleanup are performed.  This is helpful when an external controller (e.g., CAPI) manages the Kubernetes node lifecycle. By default,  k8sd will remove the Kubernetes node when it is removed from the cluster. | "true"\|"false" |

<!-- Links -->

[bootstrap]: /snap/reference/bootstrap-config-reference
