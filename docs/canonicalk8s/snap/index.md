---
myst:
  html_meta:
    description: "Canonical Kubernetes snap documentation: a lightweight, secure Kubernetes distribution installable as a snap, with built-in networking, ingress, and storage."
---

# {{product}} snap documentation

```{toctree}
:hidden:
Overview <self>
```

```{toctree}
:hidden:
:titlesonly:
:maxdepth: 6
tutorial/index.md
howto/index.md
explanation/index.md
reference/index.md
```

The {{product}} snap is a performant, lightweight, secure and
opinionated distribution of **Kubernetes** which includes everything needed to
create and manage a scalable cluster suitable for all use cases.

You can find out more about {{product}} on the 
[what is Canonical Kubernetes page] or see a more detailed explanation in our
[architecture documentation].

For deployment at scale, {{product}} is also available as a
[Juju charm][]

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

[what is Canonical Kubernetes page]: /about
[architecture documentation]: /snap/explanation/architecture
[Juju charm]: /charm/index
[Diátaxis framework]: https://diataxis.fr/
