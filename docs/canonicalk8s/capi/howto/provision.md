# Provision a {{product}} cluster with CAPI

This guide covers how to deploy a {{product}} multi-node cluster
using Cluster API (CAPI).

## Prerequisites

This guide assumes the following:

- A CAPI management cluster initialized with the infrastructure, bootstrap and
  control plane providers of your choice. Please refer to the
  [getting-started guide] for instructions.

## Generate a cluster spec manifest

You can generate a cluster manifest for a selected set of commonly used
infrastructures via templates provided by the {{product}} team.
Ensure you have initialized the desired infrastructure provider and fetch
the {{product}} provider repository:

```
git clone https://github.com/canonical/cluster-api-k8s
```

Review the list of variables needed for the cluster template:

```
cd cluster-api-k8s
export CLUSTER_NAME=yourk8scluster
clusterctl generate cluster ${CLUSTER_NAME} --from ./templates/<infrastructure-provider>/cluster-template.yaml --list-variables
```

Set the respective environment variables by editing the rc file as needed
before sourcing it. Then generate the cluster manifest:

```
source ./templates/<infrastructure-provider>/template-variables.rc
clusterctl generate cluster ${CLUSTER_NAME} --from ./templates/<infrastructure-provider>/cluster-template.yaml > cluster.yaml
```

Each provisioned node is associated with a `CK8sConfig`, through which you can
set the cluster’s properties. Available configuration fields can be listed in
detail with:

```
sudo k8s kubectl explain CK8sConfig.spec
```

Review the available options in the respective
definitions file and edit the cluster manifest (`cluster.yaml` above) to match
your needs.

## Deploy the cluster

To deploy the cluster, run:

```
sudo k8s kubectl apply -f cluster.yaml
```

For an overview of the cluster status, run:

```
clusterctl describe cluster ${CLUSTER_NAME}
```

To get the list of provisioned clusters:

```
sudo k8s kubectl get clusters
```

To see the deployed machines:

```
sudo k8s kubectl get machine
```

After the first control plane node is provisioned, you can get the kubeconfig
of the workload cluster:

```
clusterctl get kubeconfig ${CLUSTER_NAME} > ./${CLUSTER_NAME}-kubeconfig
```

You can then see the workload nodes using:

```
KUBECONFIG=./${CLUSTER_NAME}-kubeconfig sudo k8s kubectl get node
```

## Delete the cluster

To delete a cluster, run:

```
sudo k8s kubectl delete cluster ${CLUSTER_NAME}
```

<!-- LINKS -->

[getting-started guide]: ../tutorial/getting-started
