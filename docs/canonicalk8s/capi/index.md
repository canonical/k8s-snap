---
myst:
  html_meta:
    description: "Explore how to deploy and manage Canonical Kubernetes with Cluster API through step-by-step tutorials, practical how-to guides, in-depth explanations, and technical reference."
---

# Installing {{product}} with Cluster API

```{toctree}
:hidden:
Overview <self>
```

```{toctree}
:hidden:
:titlesonly:
:glob:
:caption: Deploy with Cluster API
tutorial/index.md
howto/index.md
explanation/index.md
reference/index.md
```

Cluster API (CAPI) is a Kubernetes project focused on providing declarative
APIs and tooling to simplify provisioning, upgrading, and operating multiple
Kubernetes clusters. The supporting infrastructure, like virtual machines,
networks, load balancers, and VPCs, as well as the cluster configuration are
all defined in the same way that cluster operators are already familiar with.
{{product}} supports deploying and operating Kubernetes through CAPI.

## How this documentation is organized

This documentation embodies the [Diátaxis framework].

- The [Tutorial](tutorial/getting-started) takes you step-by-step through 
  deploying your first {{product}} cluster.
- [How-to guides](howto/index) provide directions covering key cluster 
  operations and common tasks.
- [Reference](reference/index) contains technical definitions of APIs, 
  configuration and internal components.
- [Explanation](explanation/index) includes topic overviews, background and 
  context and detailed discussion.

## Project and community

{{product}} is a member of the Ubuntu family. It's an open source
project which welcomes community involvement, contributions, suggestions, fixes
and constructive feedback.

### Get involved

- [Canonical Kubernetes Slack]
- [Canonical Kubernetes Discourse]
- Our [community]
- How to [contribute]

### Releases 

- Our [release notes][releases]

### Governance and policies

- Our [Code of Conduct]

### Commercial support

Thinking about using {{product}} for your next project? [Get in touch!]

<!-- IMAGES -->

[logo]: https://assets.ubuntu.com/v1/843c77b6-juju-at-a-glace.svg

<!-- LINKS -->

[Code of Conduct]: https://ubuntu.com/community/ethos/code-of-conduct
[community]: /community
[contribute]: ../snap/howto/contribute
[releases]: https://github.com/canonical/cluster-api-k8s/releases
[overview page]: /about
[Juju]: https://juju.is
[Diátaxis framework]: https://diataxis.fr/
[Canonical Kubernetes Slack]: https://kubernetes.slack.com/archives/CG1V2CAMB
[Canonical Kubernetes Discourse]: https://discourse.ubuntu.com/c/kubernetes/180
[Get in touch!]: https://ubuntu.com/kubernetes/contact-us
