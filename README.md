# Canonical Kubernetes Snap

[![End to End Tests](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml/badge.svg)](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml)
![](https://img.shields.io/badge/Kubernetes-1.32-326de6.svg)

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/k8s)

**Canonical Kubernetes** is the fastest, easiest way to deploy a
fully-conformant Kubernetes cluster. Harnessing pure upstream Kubernetes, this
distribution adds the missing pieces (e.g. ingress, dns, networking) for a
zero-ops experience.

## Basic usage

Easily interact with the built-in features such as the load balancer:

```bash
sudo k8s get load-balancer
```

Enable the features and configure them to best suit your
cluster's needs:

```bash
sudo k8s enable load-balancer
sudo k8s set load-balancer.l2-mode=true
```

Use `kubectl` to interact with k8s:

```bash
sudo k8s kubectl get pods -A
```

If you want explore more possibilities of what you can do with Canonical
Kubernetes be sure to check out our
[how-to guides](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/)
.

## Installation

Canonical Kubernetes is compatible with Ubuntu 22.04 or later. It is also
available to other operating systems that support snap packages.

Get a single node cluster installed and intialized with just two commands:

```bash
sudo snap install k8s --channel=1.32-classic/stable --classic
sudo k8s bootstrap
```

If you would like customize the deployment, see
our
[install guides](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/install/)
.

## Documentation

For more information and instructions, please see the our
[official documentation](https://documentation.ubuntu.com/canonical-kubernetes/)
.

## Community and support

Do you have questions about Canonical Kubernetes? Perhaps you’d like some advice
from more experienced users or discuss how to achieve a certain goal? There are
a number of ways to get in touch:

- Using the [Kubernetes slack](http://slack.kubernetes.io/):
find us in the #canonical-kubernetes channel
- On the [Ubuntu Discourse](https://discourse.ubuntu.com/c/kubernetes/180)

Canonical Kubernetes is covered by the
[Ubuntu Code of Conduct](https://ubuntu.com/community/ethos/code-of-conduct).

## Contribute to Canonical Kubernetes

Canonical Kubernetes is proudly an open source project and we welcome and
encourage contributions to the code and documentation. If you are interested
, take a look at our
[contributing guide](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/contribute/)
.

## License and copyright

Canonical Kubernetes is released under the [GPL-3.0 license](LICENSE).

© 2015-2025 Canonical Ltd.
