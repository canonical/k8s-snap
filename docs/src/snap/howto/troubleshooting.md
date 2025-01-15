# How to troubleshoot {{product}}

Identifying issues in a Kubernetes cluster can be difficult, especially to new users. With {{product}} we aim to make deploying and managing your cluster as easy as possible. This how-to guide will walk you through the steps to troubleshoot your {{product}} cluster.

## Common issues

Maybe your issue has already been solved? Check out the [troubleshooting reference][snap-troubleshooting-reference] page to see a list of common issues and their solutions. Otherwise continue with this guide to help troubleshoot your {{product}} cluster.

## Check the cluster status

Verify that the cluster status is ready by running the following command:

```
sudo k8s status
```

You should see a command output similar to the following:
```
cluster status:           ready
control plane nodes:      10.94.106.249:6400 (voter), 10.94.106.208:6400 (voter), 10.94.106.99:6400 (voter)
high availability:        yes
datastore:                k8s-dqlite
network:                  enabled
dns:                      enabled at 10.152.183.106
ingress:                  disabled
load-balancer:            disabled
local-storage:            enabled at /var/snap/k8s/common/rawfile-storage
gateway                   enabled
```


## Test the API server health

Verify that the API server is healthy and reachable by running the following command on a control-plane node:

```
sudo k8s kubectl get all
```

This command lists resources that exist under the default namespace. You should see a command output similar to the following if the API server is healthy:
```
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.152.183.1   <none>        443/TCP   29m
```

A typical error message may look like this if the API server can not be reached:
```
The connection to the server 127.0.0.1:6443 was refused - did you specify the right host or port?
```

A failure can mean that the API server on the particular node is unhealthy. Check the status of the API server service:
```
sudo systemctl status snap.k8s.kube-apiserver
```

Access the logs of the API server service by running the following command:
```
sudo journalctl -xe -u snap.k8s.kube-apiserver
```

If you are trying to reach the API server from a host that is not a control-plane node, a failure could mean that:
* The API server is not reachable due to network issues or firewall limitations
* The API server is failing on the control-plane node that's being reached
* The control-plane node that's being reached is down

```{warning}
When running `sudo k8s config` on a control-plane node you retrieve the kubeconfig file that uses this node's IP address.
```

Try reaching the API server on a different control-plane node by updating the IP address that's used in the kubeconfig file.

## Check the cluster nodes' health

Confirm that the nodes in the cluster are healthy by looking for the `Ready` status:

```
sudo k8s kubectl get nodes
```

You should see a command output similar to the following:

```
NAME     STATUS   ROLES                  AGE     VERSION
node-1   Ready    control-plane,worker   10m     v1.32.0
node-2   Ready    control-plane,worker   6m51s   v1.32.0
node-3   Ready    control-plane,worker   6m21s   v1.32.0
```

## Troubleshooting an unhealthy node

Every healthy {{ product }} node has certain services up and running. The required services depend on the type of node.

Services running on both control plane and worker nodes:
* `k8sd`
* `kubelet`
* `containerd`
* `kube-proxy`

Services running only on control-plane nodes:
* `kube-apiserver`
* `kube-controller-manager`
* `kube-scheduler`
* `k8s-dqlite`

Services running only on worker nodes:
* `k8s-apiserver-proxy`

Check the status of these services on the failing node by running the following command:

```
sudo systemctl status snap.k8s.<service>
```

The logs of a failing service can be checked by running the following command:

```
sudo journalctl -xe -u snap.k8s.<service>
```

If the issue indicates a problem with the configuration of the services on the node, examine the arguments used to run these services.

The arguments of a service on the failing node can be examined by reading the file located at `/var/snap/k8s/common/args/<service>`.

## Investigating system pods' health

Check whether all of the cluster's pods are `Running` and `Ready`:

```
sudo k8s kubectl get pods -n kube-system
```

The pods in the `kube-system` namespace belong to {{product}} features such as `network`. Unhealthy pods could be related to configuration issues or nodes not meeting certain requirements.

## Troubleshooting a failing pod

Look at the events on a failing pod by running:

```
sudo k8s kubectl describe pod <pod-name> -n <namespace>
```

Check the logs on a failing pod by running the following command:

```
sudo k8s kubectl logs <pod-name> -n <namespace>
```

You can check out the upstream [debug pods documentation][] for more information.

## Using the built-in inspection script

{{product}} ships with a script to compile a complete report on {{product}} and its underlying system. This is an essential tool for bug reports and for investigating whether a system is (or isn’t) working.

Run the inspection script, by entering the command (admin privileges are required to collect all the data):

```
sudo /snap/k8s/current/k8s/scripts/inspect.sh
```

The command output is similar to the following:
```
Collecting service information
Running inspection on a control-plane node
 INFO:  Service k8s.containerd is running
 INFO:  Service k8s.kube-proxy is running
 INFO:  Service k8s.k8s-dqlite is running
 INFO:  Service k8s.k8sd is running
 INFO:  Service k8s.kube-apiserver is running
 INFO:  Service k8s.kube-controller-manager is running
 INFO:  Service k8s.kube-scheduler is running
 INFO:  Service k8s.kubelet is running
Collecting registry mirror logs
Collecting service arguments
 INFO:  Copy service args to the final report tarball
Collecting k8s cluster-info
 INFO:  Copy k8s cluster-info dump to the final report tarball
Collecting SBOM
 INFO:  Copy SBOM to the final report tarball
Collecting system information
 INFO:  Copy uname to the final report tarball
 INFO:  Copy snap diagnostics to the final report tarball
 INFO:  Copy k8s diagnostics to the final report tarball
Collecting networking information
 INFO:  Copy network diagnostics to the final report tarball
Building the report tarball
 SUCCESS:  Report tarball is at /root/inspection-report-20250109_132806.tar.gz
```

Use the report to ensure that all necessary services are running and dive into every aspect of the system.

## Reporting a bug
If you cannot solve your issue and believe that the fault may lie in {{product}}, please [file an issue on the project repository][].

Help us deal effectively with issues by including the report obtained from the inspect script, any additional logs, and a summary of the issue.

You can check out the upstream [debug documentation][] for more details on troubleshooting a Kubernetes cluster.

<!-- Links -->

[file an issue on the project repository]: https://github.com/canonical/k8s-snap/issues/new/choose
[snap-troubleshooting-reference]: ../reference/troubleshooting
[debug pods documentation]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods
[debug documentation]: https://kubernetes.io/docs/tasks/debug
