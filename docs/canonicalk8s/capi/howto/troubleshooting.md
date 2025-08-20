# How to troubleshoot {{product}}

Identifying issues in a Kubernetes cluster can be difficult, especially to new
users. With {{product}} we aim to make deploying and managing your cluster as
easy as possible. This how-to guide will walk you through the steps to
troubleshoot your {{product}} cluster.

## Check the cluster status

Verify that the cluster status is ready by running:

```
sudo k8s kubectl get cluster,ck8scontrolplane,machinedeployment,machine
```

You should see a command output similar to the following:

```
NAME                                  CLUSTERCLASS   PHASE         AGE   VERSION
cluster.cluster.x-k8s.io/my-cluster                  Provisioned   16m

NAME                                                                      INITIALIZED   API SERVER AVAILABLE   VERSION   REPLICAS   READY   UPDATED   UNAVAILABLE
ck8scontrolplane.controlplane.cluster.x-k8s.io/my-cluster-control-plane   true          true                   v1.32.1   1          1       1

NAME                                                        CLUSTER      REPLICAS   READY   UPDATED   UNAVAILABLE   PHASE     AGE   VERSION
machinedeployment.cluster.x-k8s.io/my-cluster-worker-md-0   my-cluster   1          1       1         0             Running   16m   v1.32.1

NAME                                                          CLUSTER      NODENAME                                           PROVIDERID      PHASE     AGE   VERSION
machine.cluster.x-k8s.io/my-cluster-control-plane-j7w6m       my-cluster   my-cluster-cp-my-cluster-control-plane-j7w6m       <provider-id>   Running   16m   v1.32.1
machine.cluster.x-k8s.io/my-cluster-worker-md-0-8zlzv-7vff7   my-cluster   my-cluster-wn-my-cluster-worker-md-0-8zlzv-7vff7   <provider-id>   Running   80s   v1.32.1
```

## Check providers status

{{product}} cluster provisioning failures could happen in multiple providers
used in CAPI.

Check the {{product}} bootstrap provider logs:

```
k8s kubectl logs -n cabpck-system deployment/cabpck-bootstrap-controller-manager
```

Examine the {{product}} control-plane provider logs:

```
k8s kubectl logs -n cacpck-system deployment/cacpck-controller-manager
```

Review the CAPI controller logs:

```
k8s kubectl logs -n capi-system deployment/capi-controller-manager
```

Check the logs for the infrastructure provider by running:

```
k8s kubectl logs -n <infrastructure-provider-namespace> <infrastructure-provider-deployment>
```

## Test the API server health

Fetch the kubeconfig file for a {{product}} cluster provisioned through CAPI by
running:

```
clusterctl get kubeconfig ${CLUSTER_NAME} > ./${CLUSTER_NAME}-kubeconfig.yaml
```

Verify that the API server is healthy and reachable by running:

```
kubectl --kubeconfig ${CLUSTER_NAME}-kubeconfig.yaml get all
```

This command lists resources that exist under the default namespace. If the API
server is healthy you should see a command output similar to the following:

```
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.152.183.1   <none>        443/TCP   29m
```

A typical error message may look like this if the API server can not be reached:

```
The connection to the server 127.0.0.1:6443 was refused - did you specify the right host or port?
```

A failure can mean that:

* The API server is not reachable due to network issues or firewall limitations
* The API server on the particular node is unhealthy
* All control plane nodes are down

## Check the cluster nodes' health

Confirm that the nodes in the cluster are healthy by looking for the `Ready`
status:

```
kubectl --kubeconfig ${CLUSTER_NAME}-kubeconfig.yaml get nodes
```

You should see a command output similar to the following:

```
NAME                                               STATUS   ROLES                  AGE     VERSION
my-cluster-cp-my-cluster-control-plane-j7w6m       Ready    control-plane,worker   17m     v1.32.1
my-cluster-wn-my-cluster-worker-md-0-8zlzv-7vff7   Ready    worker                 2m14s   v1.32.1
```

## Troubleshoot an unhealthy node

Every healthy {{ product }} node has certain services up and running. The
required services depend on the type of node.

Services running on both the control plane and worker nodes:

* `k8sd`
* `kubelet`
* `containerd`
* `kube-proxy`

Services running only on the control-plane nodes:

* `kube-apiserver`
* `kube-controller-manager`
* `kube-scheduler`
* `etcd`

Services running only on the worker nodes:

* `k8s-apiserver-proxy`

Make the necessary adjustments for SSH access depending on your infrastructure
provider and SSH into the unhealthy node with:

```
ssh <user>@<node>
```

Check the status of the services on the failing node by running:

```
sudo systemctl status snap.k8s.<service>
```

Check the logs of a failing service by executing:

```
sudo journalctl -xe -u snap.k8s.<service>
```

If the issue indicates a problem with the configuration of the services on the
node, examine the arguments used to run these services.

The arguments of a service on the failing node can be examined by reading the
file located at `/var/snap/k8s/common/args/<service>`.

## Investigate system pods' health

Check whether all of the cluster's pods are `Running` and `Ready`:

```
kubectl --kubeconfig ${CLUSTER_NAME}-kubeconfig.yaml get pods -n kube-system
```

The pods in the `kube-system` namespace belong to {{product}}' features such as
`network`. Unhealthy pods could be related to configuration issues or nodes not
meeting certain requirements.

## Troubleshoot a failing pod

Look at the events on a failing pod by running:

```
kubectl --kubeconfig ${CLUSTER_NAME}-kubeconfig.yaml describe pod <pod-name> -n <namespace>
```

Check the logs on a failing pod by executing:

```
kubectl --kubeconfig ${CLUSTER_NAME}-kubeconfig.yaml logs <pod-name> -n <namespace>
```

You can check out the upstream [debug pods documentation][] for more
information.

## Use the built-in inspection script

{{product}} ships with a script to compile a complete report on {{product}} and
its underlying system. This is an essential tool for bug reports and for
investigating whether a system is (or isn’t) working.

The inspection script can be executed on a specific node by running the
following commands:

```
ssh -t <user>@<node> -- sudo k8s inspect /home/<user>/inspection-report.tar.gz
scp <user>@<node>:/home/<user>/inspection-report.tar.gz ./
```

See the [inspection report reference page] for more details.

## Report a bug

If you cannot solve your issue and believe that the fault may lie in
{{product}}, please [file an issue on the project repository][].

Help us deal effectively with issues by including the report obtained from the
inspect script, any additional logs, and a summary of the issue.

You can check out the upstream [debug documentation][] for more details on
troubleshooting a Kubernetes cluster.

<!-- Links -->

[file an issue on the project repository]: https://github.com/canonical/cluster-api-k8s/issues/new/choose
[capi-troubleshooting-reference]: ../reference/troubleshooting
[systemd]: https://systemd.io
[debug pods documentation]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods
[debug documentation]: https://kubernetes.io/docs/tasks/debug
[inspection report reference page]: /snap/reference/inspection-reports.md
