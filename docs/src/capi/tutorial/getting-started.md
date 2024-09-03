# Cluster provisioning with CAPI and Canonical K8s

This guide covers how to deploy a {{product}} multi-node cluster
using Cluster API (CAPI).

## Install `clusterctl`

The `clusterctl` CLI tool manages the lifecycle of a Cluster API management
cluster. To install it, follow the [upstream instructions]. Typically, this
involves fetching the executable that matches your hardware architecture and
placing it in your PATH. For example, at the time this guide was written,
for `amd64` you would run:

```
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.7.3/clusterctl-linux-amd64 -o clusterctl
sudo install -o root -g root -m 0755 clusterctl /usr/local/bin/clusterctl
```

### Configure clusterctl

`clusterctl` contains a list of default providers. Right now, {{product}} is 
not yet part of that list. To make `clusterctl` aware of the new
providers we need to add them to the configuration
file. Edit `~/.cluster-api/clusterctl.yaml` and add the following:

```
providers:
  - name: ck8s
    type: BootstrapProvider
    url: "https://github.com/canonical/cluster-api-k8s/releases/latest/bootstrap-components.yaml"
  - name: ck8s
    type: ControlPlaneProvider
    url: "https://github.com/canonical/cluster-api-k8s/releases/latest/control-plane-components.yaml"
    type: "ControlPlaneProvider"
```

### Set up a management cluster

The management cluster hosts the CAPI providers. You can use Canonical
Kubernetes as a management cluster:

```
sudo snap install k8s --classic --edge
sudo k8s bootstrap
sudo k8s status --wait-ready
mkdir -p ~/.kube/
sudo k8s config > ~/.kube/config
```

When setting up the management cluster, place its kubeconfig under
`~/.kube/config` so other tools such as `clusterctl` can discover and interact
with it.

### Prepare the infrastructure provider

Before generating a cluster, you need to configure the infrastructure provider.
Each provider has its own prerequisites. Please follow the instructions
for your provider:

`````{tabs}
````{group-tab} AWS

The AWS infrastructure provider requires the `clusterawsadm` tool to be
installed:

```
curl -L https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/download/v2.5.2/clusterawsadm-linux-amd64 -o clusterawsadm
chmod +x clusterawsadm
sudo mv clusterawsadm /usr/local/bin
```

`clusterawsadm` helps you bootstrapping the AWS environment that CAPI will use
It will also create the necessary IAM roles for you.

Start by setting up environment variables defining the AWS account to use, if
these are not already defined:

```
export AWS_REGION=<your-region-eg-us-east-1>
export AWS_ACCESS_KEY_ID=<your-access-key>
export AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
```

If you are using multi-factor authentication, you will also need:

```
export AWS_SESSION_TOKEN=<session-token>
```

`clusterawsadm` uses these details to create a [CloudFormation] stack in your
AWS account with the correct [IAM] resources:

```
clusterawsadm bootstrap iam create-cloudformation-stack
```

The credentials need to be encoded and stored as a Kubernetes secret:

```
export AWS_B64ENCODED_CREDENTIALS=$(clusterawsadm bootstrap credentials encode-as-profile)
```

You are now all set to deploy the AWS CAPI infrastructure provider.
````

````{group-tab} MAAS
Start by setting up environment variables to allow access to MAAS:

```
export MAAS_API_KEY="<maas-api-key>"
export MAAS_ENDPOINT="http://<maas-endpoint>/MAAS"
export MAAS_DNS_DOMAIN="<maas-dns-domain>"
```
The MAAS infrastructure provider uses these credentials to deploy machines,
create DNS records and perform various other operations for workload clusters.

```{warning}
The management cluster needs to resolve DNS records from the MAAS domain, 
therefore it should be deployed on a MAAS machine.
```

Define further environment variables for the machine image and minimum compute
resources of the control plane and worker nodes:

```
export CONTROL_PLANE_MACHINE_MINCPU="1"
export CONTROL_PLANE_MACHINE_MINMEMORY="2048"
export CONTROL_PLANE_MACHINE_IMAGE="ubuntu"

export WORKER_MACHINE_MINCPU="1"
export WORKER_MACHINE_MINMEMORY="2048"
export WORKER_MACHINE_IMAGE="ubuntu"
```

```{note}
The minimum resource variables are used to select machines with resources more
than or equal to the provided values.
```

Optional environment variables can be defined for specifying resource pools
and machine tags:

```
# (optional) Configure resource pools for control plane and worker machines
# export CONTROL_PLANE_MACHINE_RESOURCEPOOL="kvm-pool"
# export WORKER_MACHINE_RESOURCEPOOL="bare-metal-pool"

# (optional) Configure (comma-separated) tags for control plane and worker machines
# export CONTROL_PLANE_MACHINE_TAGS="control-plane,controller"
# export WORKER_MACHINE_TAGS="worker,compute"
```

You are now all set to deploy the MAAS CAPI infrastructure provider.
````
`````

### Initialise the management cluster

To initialise the management cluster with the latest released version of the
providers and the infrastructure of your choice:

```
clusterctl init --bootstrap ck8s --control-plane ck8s -i <infra-provider-of-choice>
```

### Generate a cluster spec manifest

Once the bootstrap and control-plane controllers are up and running, you can
apply the cluster manifests with the specifications of the cluster you want to
provision.

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
set the cluster’s properties. Review the available options in the respective
definitions file and edit the cluster manifest (`cluster.yaml` above) to match
your needs.

### Deploy the cluster

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
clusterctl get kubeconfig ${CLUSTER_NAME} ${CLUSTER_NAME}-kubeconfig
```

You can then see the workload nodes using:

```
KUBECONFIG=./kubeconfig sudo k8s kubectl get node
```

### Delete the cluster

To delete a cluster:

```
sudo k8s kubectl delete cluster ${CLUSTER_NAME}
```

<!-- Links -->
[upstream instructions]: https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl
[CloudFormation]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html
[IAM]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
