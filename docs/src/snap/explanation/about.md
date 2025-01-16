# What is {{product}}?

At its core, {{product}} is a full implementation of upstream
[Kubernetes] delivered in a compact, secure, reliable [snap] package. As the
upstream Kubernetes services are not all that is required for a fully
functional cluster, additional services and features are built in. You can
deploy the snap and have a single-node cluster up and running in minutes.

## Why a snap?

Snaps are self-contained, simple to install, secure, cross-platform, and
dependency-free. They can be installed on any Linux system which supports the
`snapd` service (see the [snapd documentation] for more information). Security
and robustness are their key features, alongside being easy to install, easy to
maintain and easy to upgrade.

## What else comes with it?

In addition to the upstream Kubernetes services,
{{product}} also includes:

- a DNS service for the node
- a CNI for the node/cluster
- a simple storage provider
- an ingress provider
- a load-balancer
- a gateway API controller
- a metrics server

## Where can I install it?

The {{product}} snap can be installed on a Linux OS, wherever it may be: run it
in several local containers or VMs for example, or use it on public/private
cloud instances. For deploying with [Juju], a machine [charm] to deploy the
snap is also available.
For deployment using Cluster API, a bootstrap and
control-plane provider have also been created. See the [choosing an installation
method] page to see which method best suits your project's needs.

## Can I use it to make a cluster?

Yes. {{product}} is designed to be eminently scalable. You can start
with a single node and add more as and when the need arises. Scale up or down
at any time. Systems with more than three nodes will automatically become
Highly Available.

## Does it come with support?

Each and every user will be supported by the community. For a more detailed
look at what that entails, please see our [Community page]. If you need a
greater level of support, Canonical provides [Ubuntu Pro], a comprehensive
subscription for your open-source software stack. For more support options,
visit the [Ubuntu support] page.

## Next steps

- Try it now! Jump over to the [Getting started tutorial][tutorial]

<!-- LINKS -->

[Kubernetes]: https://kubernetes.io
[snap]: https://snapcraft.io/docs
[tutorial]: /src/snap/tutorial/getting-started.md
[Juju]: https://juju.is
[charm]: https://charmhub.io/k8s
[snapd documentation]: https://snapcraft.io/docs/installing-snapd
[Community page]: community.md
[Ubuntu Pro]:  https://ubuntu.com/pro
[Ubuntu support]: https://ubuntu.com/support
[choosing an installation method]: /src/snap/explanation/installation-methods.md
