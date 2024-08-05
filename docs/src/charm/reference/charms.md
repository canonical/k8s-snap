# {{product}} charms

{{product}} can be deployed via [Juju][] with the following charms:

- **k8s**, which deploys a complete Kubernetes implementation
- **k8s-worker**, which deploys Kubernetes without the control plane for units
  intended as workers in a cluster (at least one `k8s` charm must be deployed
  and integrated with the worker for it to function)

## Charmhub

Both of the above charms are published to Charmhub.

- [The Charmhub page for the `k8s` charm][cs-k8s]
- [The Charmhub page for the `k8s-worker` charm][cs-k8s-worker]

For an explanation of the releases and channels, please see the documentation
[explaining channels][].


## Source

The source code for both charms is contained in a single repository:

[https://github.com/canonical/k8s-operator][repo]

Please see the [readme file][] there for further specifics of the charm
implementation.

<!-- LINKS -->
[Juju]: https://juju.is
[explaining channels]: /charm/explanation/channels
[cs-k8s]: https://charmhub.io/k8s
[cs-k8s-worker]: https://charmhub.io/k8s-worker
[readme file]: https://github.com/canonical/k8s-operator#readme
[repo]: https://github.com/canonical/k8s-operator