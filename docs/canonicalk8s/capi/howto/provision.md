---
myst:
  html_meta:
    description: Set up a Cluster API management cluster, then provision, deploy and delete a multi-node Canonical Kubernetes cluster in this how-to guide.
relatedlinks: "[Cluster&#32;API&#32;documentation](https://cluster-api.sigs.k8s.io/), [clusterctl&#32;installation](https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl), [Cluster&#32;API&#32;Provider&#32;MAAS](https://github.com/spectrocloud-public/cluster-api-provider-maas)"
---

# Provision a {{product}} cluster with CAPI

This guide covers how to set up a Cluster API (CAPI) management cluster and
use it to deploy a {{product}} multi-node cluster.

## Prerequisites

This guide assumes the following:

- [`clusterctl`](https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl) installed.

## Set up a management cluster

The management cluster hosts the CAPI providers and manages the lifecycle of
your workload clusters. Install {{product}}:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
:append: sudo k8s bootstrap
```

Place the kubeconfig of the management cluster under `~/.kube/config` so that
`clusterctl` can discover it:

```
sudo k8s status --wait-ready
mkdir -p ~/.kube/
sudo k8s config > ~/.kube/config
```

If you use MAAS as the infrastructure provider, deploy the management cluster
on a MAAS machine so that it can resolve DNS records from the MAAS domain.

## Prepare the infrastructure provider

Each infrastructure provider has its own prerequisites that must be met before
you initialize the management cluster, usually involving credentials that you
export as environment variables.

{{product}} provides [cluster templates](https://github.com/canonical/cluster-api-k8s/tree/main/templates)
for a selected set of infrastructures, such as `aws` and `maas`. Choose an
infrastructure provider that has a template, because you use it later to
generate the cluster spec manifest. Your choice is referred to as
`<infrastructure-provider>` in commands for the remainder of this guide, and
must match the name of its template directory.

Use the [list of Cluster API providers](https://cluster-api.sigs.k8s.io/reference/providers#infrastructure)
to find the documentation of your provider. Cluster API provides a [quick start](https://cluster-api.sigs.k8s.io/user/quick-start#initialization-for-common-providers) guide to help prepare your environment. Only follow the preparation steps in the provider instructions.

```{important}
Do not run the `clusterctl init` command shown in the provider instructions.
Once your environment is ready, return to this guide and initialize the
management cluster with the {{product}} providers, as described in
{ref}`Initialize the management cluster <initialize-management-cluster>`.
```

(initialize-management-cluster)=

## Initialize the management cluster

Initialize the management cluster with the {{product}} bootstrap and control
plane providers and the infrastructure provider of your choice:

```
clusterctl init --bootstrap canonical-kubernetes --control-plane canonical-kubernetes --infrastructure <infrastructure-provider>
```

For AWS, run the following command instead. It pins the CAPI core and CAPA
versions to avoid a CRD version mismatch:

```
clusterctl init --core cluster-api:v1.9.6 --bootstrap canonical-kubernetes --control-plane canonical-kubernetes --infrastructure aws:v2.8.4
```

Confirm that the provider controllers are running:

```
sudo k8s kubectl get pods --all-namespaces
```

Wait until all the pods in the provider namespaces report a `Running` status
before you continue.

(generate-cluster-spec)=

## Generate a cluster spec manifest

Generate a cluster manifest from the {{product}} template for your
infrastructure provider. The templates are in the {{product}} provider
repository, which you fetch with:

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

If the deployment fails, see the [troubleshooting guide](troubleshooting.md).

## Delete the cluster

To delete a cluster, run:

```
sudo k8s kubectl delete cluster ${CLUSTER_NAME}
```
