# Cluster provisioning with CAPI and Canonical K8s

This guide covers how to deploy a Canonical Kubernetes multi-node cluster using Cluster API (CAPI).

## Install `clusterctl`

The `clusterctl` CLI tool manages the lifecycle of a Cluster API management cluster. To install it, follow the [upstream instructions]. Typically, this involves fetching the executable that matches your hardware architecture and placing it in your PATH. For example, at the time this guide was written, for `amd64` you would run:

```sh
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.7.3/clusterctl-linux-amd64 -o clusterctl
sudo install -o root -g root -m 0755 clusterctl /usr/local/bin/clusterctl
```

### Configure clusterctl

`clusterctl` contains a list of default providers. Right now, Canonical Kubernetes is not yet part of that list. To make `clusterctl` aware of the Canonical K8s providers, we need to add a clusterctl configuration file.

```sh
mkdir -p ~/.config/cluster-api
curl -L https://raw.githubusercontent.com/canonical/cluster-api-k8s/main/clusterctl.yaml -o ~/.config/cluster-api/clusterctl.yaml
```

### Set up a management cluster

The management cluster hosts the CAPI providers. You can use Canonical Kubernetes  as a management cluster:

```sh
sudo snap install k8s --classic --edge
sudo k8s bootstrap
sudo k8s status --wait-ready
mkdir -p ~/.kube/
sudo k8s config > ~/.kube/config
```

When setting up the management cluster, place its kubeconfig under `~/.kube/config` so other tools such as `clusterctl` can discover and interact with it.

### Prepare the Infrastructure Provider

Before generating a cluster, you need to configure the infrastructure provider. Each provider has its own prerequisites. Please follow the instructions for your provider:

* [aws][aws-provider]
<!-- TO BE EXTENDED -->

### Initialize the Management Cluster

To initialize the management cluster with the latest released version of the providers and the infrastructure of your choice:

```sh
clusterctl init --bootstrap ck8s --control-plane ck8s -i <infra-provider-of-choice>
```

### Generate a Cluster Spec Manifest

Once the bootstrap and control-plane controllers are up and running, you can apply the cluster manifests with the specifications of the cluster you want to provision.

You can generate a cluster manifest for a selected set of commonly used infrastructures via templates provided by the Canonical Kubernetes team. Ensure you have initialized the desired infrastructure provider and fetch the Canonical Kubernetes provider repository:

```sh
git clone https://github.com/canonical/cluster-api-k8s
```

Review the list of variables needed for the cluster template:

```sh
cd cluster-api-k8s
clusterctl generate cluster <cluster-name> --from ./templates/<infrastructure-provider>/cluster-template.yaml --list-variables
```

Set the respective environment variables by editing the rc file as needed before sourcing it. Then generate the cluster manifest:

```sh
source ./templates/<infrastructure-provider>/template-variables.rc
clusterctl generate cluster <cluster-name> --from ./templates/<infrastructure-provider>/cluster-template.yaml > cluster.yaml
```

Each provisioned node is associated with a `CK8sConfig`, through which you can set the cluster’s properties. Review the available options in the respective definitions file and edit the cluster manifest (`cluster.yaml` above) to match your needs.

### Deploy the Cluster

To deploy the cluster, run:

```sh
sudo k8s kubectl apply -f cluster.yaml
```

For a overview of the cluster status, run:

```sh
clusterctl describe cluster <cluster-name>
```

To get the list of provisioned clusters:

```sh
sudo k8s kubectl get clusters
```

To see the deployed machines:

```sh
sudo k8s kubectl get machine
```

After the first control plane node is provisioned, you can get the kubeconfig of the workload cluster:

```sh
clusterctl get kubeconfig <cluster-name> > kubeconfig
```

You can then see the workload nodes using:

```sh
KUBECONFIG=./kubeconfig sudo k8s kubectl get node
```

### Delete the Cluster

To delete a cluster:

```sh
sudo k8s kubectl delete cluster <cluster-name>
```

<!-- Links -->
[upstream instructions]: https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl
[aws-provider]: ../howto/aws-provider.md
