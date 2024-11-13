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

Cluster API (CAPI) is a Kubernetes project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters. The supporting infrastructure, like virtual machines, networks, load balancers, and VPCs, as well as the cluster configuration are all defined in the same way that cluster operators are already familiar with. {{product}} supports deploying and operating Kubernetes through CAPI.

![Illustration depicting working on components and clouds][logo]

## In this documentation

````{grid} 1 1 2 2

```{grid-item-card} [Tutorial](tutorial/index)

**Start here!** A hands-on introduction to {{product}} for new users
```

```{grid-item-card} [How-to guides](howto/index)

**Step-by-step guides** covering key operations and common tasks
```

````

````{grid} 1 1 2 2


```{grid-item-card} [Reference](reference/index)

**Technical information** - specifications, APIs, architecture
```

```{grid-item-card} [Explanation](explanation/index)

**Discussion and clarification** of key topics
```

````

---

## Project and community

{{product}} is a member of the Ubuntu family. It's an open source
project which welcomes community involvement, contributions, suggestions, fixes
and constructive feedback.

- Our [Code of Conduct]
- Our [community]
- How to [contribute]
- Our development [roadmap]

<!-- IMAGES -->

[logo]: https://assets.ubuntu.com/v1/843c77b6-juju-at-a-glace.svg

<!-- LINKS -->

[Code of Conduct]: https://ubuntu.com/community/ethos/code-of-conduct
[community]: ../charm/reference/community
[contribute]: ../snap/howto/contribute
[roadmap]: ../snap/reference/roadmap
[overview page]: ../charm/explanation/about
[arch]: ../charm/reference/architecture
[Juju]: https://juju.is
[k8s snap package]: ../snap/index