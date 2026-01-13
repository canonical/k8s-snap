# Canonical Kubernetes Snap

[![End to End Tests](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml/badge.svg)](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml)
[![](https://snapcraft.io/k8s/badge.svg)](https://snapcraft.io/k8s)

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/k8s)

**Canonical Kubernetes** is an opinionated distribution of Kubernetes which
includes all the tools needed to create and manage a scalable cluster with
[LTS](https://canonical.com/blog/12-year-lts-for-kubernetes). Canonical
Kubernetes builds on the main Kubernetes project by providing all
the necessary pieces for a zero-ops experience, such as Ingress, DNS,
networking, and so on. Whether you are a complete beginner to Kubernetes or a
seasoned system administrator, Canonical Kubernetes provides a way to easily
deploy a cluster allowing you to focus on applications over infrastructure.

## Basic usage

Canonical Kubernetes provides a way for you to easily enable, disable, and
configure the essential default Kubernetes features in your cluster.

For example, if you want a load balancer with L2 mode enabled, run:

```bash
sudo k8s enable load-balancer
sudo k8s set load-balancer.l2-mode=true
```

Or, if you want to disable the default local storage before implementing your
own storage solution, run:

```bash
sudo k8s disable local-storage
```

Use kubectl to interact with k8s just as you would with any other Kubernetes
cluster:

```bash
sudo k8s kubectl get pods -A
```

If you want to explore the possibilities of what you can do with Canonical
Kubernetes, be sure to check out its
[how-to guides](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/).

## Installation

Canonical Kubernetes is available for Ubuntu 22.04 and higher. It's also
available in other Linux distributions that support
[snaps](https://snapcraft.io/).

Install the snap:

```bash
sudo snap install k8s --channel=1.35-classic/stable --classic
```

Initialize the cluster:

```bash
sudo k8s bootstrap
```

If you would like to customize the deployment, see our
[installation guides](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/install/).

## Documentation

The
[Canonical Kubernetes documentation](https://documentation.ubuntu.com/canonical-kubernetes/)
provides information about how to grow your cluster by adding additional nodes,
how your cluster can stay up-to-date with the latest Kubernetes releases
automatically, backing up your cluster, and much more.

## Community and support

Do you have questions about Canonical Kubernetes? Perhaps you'd like some advice
from more experienced users or discuss how to achieve a certain goal? Get in
touch on the
[#canonical-kubernetes](https://kubernetes.slack.com/archives/CG1V2CAMB)
channel on the [Kubenetes Slack workspace](http://slack.kubernetes.io/).

You can report any bugs or issues you find on
[GitHub](https://github.com/canonical/k8s-snap/issues).

Canonical Kubernetes is covered by the
[Ubuntu Code of Conduct](https://ubuntu.com/community/ethos/code-of-conduct).

## Contribute to Canonical Kubernetes

Canonical Kubernetes is a proudly open source project, and we welcome and
encourage contributions to the code and documentation. If you are interested,
take a look at our
[contributing guide](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/contribute/)
.

## License and copyright

Canonical Kubernetes is released under the [GPL-3.0 license](LICENSE).

© 2015-2026 Canonical Ltd.
