---
myst:
  html_meta:
    description: "Canonical Kubernetes charm documentation: the k8s Juju operator for deploying, scaling, and managing Kubernetes on Juju clouds."
---

# {{product}} charm documentation

```{toctree}
:hidden:
Overview <self>
```

```{toctree}
:hidden:
:titlesonly:
:caption: Deploy with Juju
tutorial/index.md
howto/index.md
explanation/index.md
reference/index.md
```

The {{product}} charm, `k8s`, is an operator-software which wraps an
application and contains all of the instructions necessary for deploying,
configuring, scaling, integrating the application on any cloud supported by
[Juju][].

The `k8s` charm takes care of installing and configuring the [k8s snap
package][] on cloud instances managed by Juju. Operating Kubernetes through
this charm makes it significantly easier to manage at scale, on remote cloud
instances and also to integrate other operators to enhance or customize your
Kubernetes deployment. You can find out more about {{product}} on the
[what is Canonical Kubernetes page][] or see a more detailed explanation in our 
[architecture documentation][arch].

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
[arch]: /charm/explanation/architecture
[Juju]: https://juju.is
[k8s snap package]: /snap/index
[Diátaxis framework]: https://diataxis.fr/