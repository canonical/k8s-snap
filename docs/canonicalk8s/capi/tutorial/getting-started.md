# Getting started with Cluster API

This guide covers how to deploy a {{product}} management cluster for Cluster
API (CAPI).

## Install `clusterctl`

The `clusterctl` CLI tool manages the lifecycle of a Cluster API management
cluster. To install it, follow the [upstream instructions]. Typically, this
involves fetching the executable that matches your hardware architecture and
placing it in your PATH. For example, at the time this guide was written,
for `amd64` you would run:

```
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.9.3/clusterctl-linux-amd64 -o clusterctl
sudo install -o root -g root -m 0755 clusterctl /usr/local/bin/clusterctl
```

For more `clusterctl` versions refer to the
[upstream release page][clusterctl-release-page].

## Set up a management cluster

The management cluster hosts the CAPI providers. You can use {{product}} as a
management cluster:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
:append: sudo k8s bootstrap
```

When setting up the management cluster, place its kubeconfig under
`~/.kube/config` so other tools such as `clusterctl` can discover and interact
with it.

```
sudo k8s status --wait-ready
mkdir -p ~/.kube/
sudo k8s config > ~/.kube/config
```

## Prepare the infrastructure provider

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

`clusterawsadm` helps you bootstrapping the AWS environment that CAPI will use.
It will also create the necessary IAM roles for you. For more `clusterawsadm`
versions refer to the [upstream release page][clusterawsadm-release-page].

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

## Initialize the management cluster

To initialize the management cluster with the latest released version of the
providers and the infrastructure of your choice:

```
clusterctl init --bootstrap canonical-kubernetes --control-plane canonical-kubernetes -i <infra-provider-of-choice>
```

Once the bootstrap and control-plane controllers are up and running, you can
apply the cluster manifests with the specifications of the cluster you want to
provision.

## Next steps

- Learn how to provision a {{product}} cluster with CAPI:
[Provision a Canonical Kubernetes cluster]
- Learn how to upgrade the providers:
[Upgrade the providers of a management cluster]
- Learn how to install a custom {{product}} version:
[Install custom Canonical Kubernetes]

<!-- Links -->
[upstream instructions]: https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl
[CloudFormation]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html
[IAM]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
[clusterctl-release-page]: https://github.com/kubernetes-sigs/cluster-api/releases
[clusterawsadm-release-page]: https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases
[Provision a Canonical Kubernetes cluster]: ../howto/provision.md
[Install custom Canonical Kubernetes]: ../howto/custom-ck8s.md
[Upgrade the providers of a management cluster]: ../howto/upgrade-providers.md
