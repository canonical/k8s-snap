# Choosing an installation method

{{ product }} can be installed in a variety of ways, depending on your needs and
preferences. All installation methods provide a fully functional cluster with
upstream Kubernetes and additional built-in features such as networking, ingress
and storage. Your choice may be influenced by the cluster size, the 
environment you are deploying to, and the life cycle management tools that you
prefer to use. The following sections describe the different installation
methods:

## Snap

The {{ product }} snap is a self-contained, simple to install package. It can
be installed on any Linux distribution that supports the
[snapd] service, such as
[Ubuntu]. If you're using a non-Linux system, we recommend creating virtual
machines using [Multipass] or [LXD]. [Snaps] come with the advantage of
automatic and atomic updates making it easy for users to install, maintain and
update their software.

If you are looking for a straightforward way to get started with {{ product }},
and your cluster will not grow beyond a few nodes, using the snap is
the recommended installation method. Follow the [getting-started guide] to
install {{ product }} using the snap.

## Juju

[Juju] is an open-source orchestration engine that allows you to
deploy, configure, scale and operate {{ product }} on any infrastructure. Juju
uses charms which are encapsulated reusable software packages to define how
applications are deployed and related to each other. At its core, {{ product }}
comprises of two Juju charms: a [control plane charm] and a [worker charm].
Additionally, the Juju charm ecosystem provides further integrations, for
example an observability stack.

If you are looking to deploy {{ product }} in a public/private cloud
environment, on metal, in VMs or on your local machine you can take advantage of
Juju's powerful lifecycle management. Get started with a simple deployment of
the {{ product }}'s charms using the [juju-cli guide] or leverage
[Terraform's Juju provider] by following the [installing-with-terraform]
guide.

## Cluster API (CAPI)

[Cluster API] is a Kubernetes sub-project that provides declarative APIs for
creating, configuring, and managing multiple Kubernetes clusters. 

If you plan to deploy and operate a large deployment with multiple
Kubernetes clusters, you can use the {{ product }}'s CAPI providers. Follow the
[CAPI guide] to deploy and operate {{ product }} with
the help of the CAPI providers.

<!-- LINKS -->

[Ubuntu]: https://help.ubuntu.com/
[Snaps]: https://snapcraft.io/docs
[snapd]: https://snapcraft.io/docs/installing-snapd
[Multipass]: https://canonical.com/multipass
[LXD]: https://canonical.com/lxd
[Juju]: https://juju.is
[juju-cli guide]: /src/charm/tutorial/getting-started.md
[control plane charm]: https://charmhub.io/k8s
[worker charm]: https://charmhub.io/k8s-worker
[getting-started guide]: /src/snap/tutorial/getting-started.md
[Terraform's Juju provider]: https://github.com/juju/terraform-provider-juju/
[installing-with-terraform]: /src/charm/howto/install-terraform.md
[CAPI guide]: /src/capi/tutorial/getting-started.md
[Cluster API]: https://cluster-api.sigs.k8s.io/
