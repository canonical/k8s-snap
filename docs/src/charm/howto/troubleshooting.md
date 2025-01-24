# How to troubleshoot {{product}}

Identifying issues in a Kubernetes cluster can be difficult, especially to new
users. With {{product}} we aim to make deploying and managing your cluster as
easy as possible. This how-to guide will walk you through the steps to
troubleshoot your {{product}} cluster.

## Common issues

Maybe your issue has already been solved? Check out the
[troubleshooting reference][charm-troubleshooting-reference] page to see a list
of common issues and their solutions. Otherwise continue with this guide to
help troubleshoot your {{product}} cluster.

## Check the cluster status

Verify that the cluster status is ready by running:

```
juju status
```

You should see a command output similar to the following:

```
Model        Controller           Cloud/Region         Version  SLA          Timestamp
k8s-testing  localhost-localhost  localhost/localhost  3.6.1    unsupported  09:06:50Z

App         Version  Status  Scale  Charm       Channel    Rev  Exposed  Message
k8s         1.32.0   active      1  k8s         1.32/beta  179  no       Ready
k8s-worker  1.32.0   active      1  k8s-worker  1.32/beta  180  no       Ready

Unit           Workload  Agent  Machine  Public address  Ports     Message
k8s-worker/0*  active    idle   1        10.94.106.154             Ready
k8s/0*         active    idle   0        10.94.106.136   6443/tcp  Ready

Machine  State    Address        Inst id        Base          AZ  Message
0        started  10.94.106.136  juju-380ff2-0  ubuntu@24.04      Running
1        started  10.94.106.154  juju-380ff2-1  ubuntu@24.04      Running
```

Interpreting the Output:

- The `Workload` column shows the status of a given service.
- The `Message` section details the health of a given service in the cluster.
- The `Agent` column reflects any activity of the Juju agent.

During deployment and maintenance the workload status will reflect the node's
activity. An example workload may display `maintenance` along with the message
details: `Ensuring snap installation`.

During normal cluster operation the `Workload` column reads `active`, the
`Agent` column shows `idle`, and the messages will either read `Ready` or
another descriptive term.

## Test the API server health

Fetch the kubeconfig file for a control-plane node in the cluster by running:

```
juju run k8s/leader get-kubeconfig | yq .kubeconfig > cluster-kubeconfig.yaml
```

```{warning}
When running `juju run k8s/leader get-kubeconfig` you retrieve the kubeconfig file that uses one of the unit's  public IP addresses in the kubernetes endpoint. This endpoint ip can be overriden by providing a `server` argument if the api is exposed through a load-balancer.
```

Verify that the API server is healthy and reachable by running:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get all
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

Check the status of the API server service:

```
juju exec --unit k8s/0 -- systemctl status snap.k8s.kube-apiserver
```

Access the logs of the API server service by running:

```
juju exec --unit k8s/0 -- journalctl -u snap.k8s.kube-apiserver
```

A failure can mean that:

* The API server is not reachable due to network issues or firewall limitations
* The API server on the particular node is unhealthy
* The control-plane node that's being reached is down

Try reaching the API server on a different unit by retrieving the kubeconfig
file with `juju run <k8s/unit#> get-kubeconfig`. Please replace `#` with the
desired unit's number.

## Check the cluster nodes' health

Confirm that the nodes in the cluster are healthy by looking for the `Ready`
status:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get nodes
```

You should see a command output similar to the following:

```
NAME            STATUS   ROLES                  AGE     VERSION
juju-380ff2-0   Ready    control-plane,worker   9m30s   v1.32.0
juju-380ff2-1   Ready    worker                 77s     v1.32.0
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
* `k8s-dqlite`

Services running only on the worker nodes:

* `k8s-apiserver-proxy`

SSH into the unhealthy node by running:

```
juju ssh <k8s/unit#>
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
kubectl --kubeconfig cluster-kubeconfig.yaml get pods -n kube-system
```

The pods in the `kube-system` namespace belong to {{product}}' features such as
`network`. Unhealthy pods could be related to configuration issues or nodes not
meeting certain requirements.

## Troubleshoot a failing pod

Look at the events on a failing pod by running:

```
kubectl --kubeconfig cluster-kubeconfig.yaml describe pod <pod-name> -n <namespace>
```

Check the logs on a failing pod by executing:

```
kubectl --kubeconfig cluster-kubeconfig.yaml logs <pod-name> -n <namespace>
```

You can check out the upstream [debug pods documentation][] for more
information.

## Use the built-in inspection script

{{product}} ships with a script to compile a complete report on {{product}} and
its underlying system. This is an essential tool for bug reports and for
investigating whether a system is (or isn’t) working.

Inspection script can be executed on a specific unit by running the following
commands:

```
juju exec --unit <k8s/unit#> -- sudo /snap/k8s/current/k8s/scripts/inspect.sh /home/ubuntu/inspection-report.tar.gz
juju scp <k8s/unit#>:/home/ubuntu/inspection-report.tar.gz ./
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
 SUCCESS:  Report tarball is at /home/ubuntu/inspection-report.tar.gz
```

Use the report to ensure that all necessary services are running and dive into
every aspect of the system.

## Collect debug information

To collect comprehensive debug output from your {{product}} cluster, install
and run [juju-crashdump][] on a computer that has the Juju client installed.
Please ensure that the current controller and model are pointing at your
{{product}} deployment.

```
sudo snap install juju-crashdump --classic --channel edge
juju-crashdump -a debug-layer -a config
```

Running the `juju-crashdump` script will generate a tarball of debug
information that includes [systemd][] unit status and logs, Juju logs, charm
unit data, and Kubernetes cluster information. Please include the generated
tarball when filing a bug.

## Report a bug

If you cannot solve your issue and believe that the fault may lie in
{{product}}, please [file an issue on the project repository][].

Help us deal effectively with issues by including the report obtained from the
inspect script, the tarball obtained from `juju-crashdump`, as well as any
additional logs, and a summary of the issue.

You can check out the upstream [debug documentation][] for more details on
troubleshooting a Kubernetes cluster.

<!-- Links -->

[file an issue on the project repository]: https://github.com/canonical/k8s-operator/issues/new/choose
[charm-troubleshooting-reference]: ../reference/troubleshooting
[juju-crashdump]: https://github.com/juju/juju-crashdump
[systemd]: https://systemd.io
[debug pods documentation]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods
[debug documentation]: https://kubernetes.io/docs/tasks/debug
