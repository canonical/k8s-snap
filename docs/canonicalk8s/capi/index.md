---
myst:
  html_meta:
    description: "Canonical Kubernetes with Cluster API: declarative cluster provisioning and lifecycle management across cloud and on-premises infrastructure."
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

<!-- LINKS -->

[Diátaxis framework]: https://diataxis.fr/
