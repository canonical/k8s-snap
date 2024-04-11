# Canonical Kubernetes documentation

Canonical Kubernetes is a performant, lightweight, secure and
opinionated distribution of **Kubernetes** which includes everything needed to
create and manage a scalable cluster suitable for all use cases.

You can find out more about Canonical Kubernetes on this [overview page] or
see a more detailed explanation in our [architecture documentation].

![Illustration depicting working on components and clouds][logo]

```{toctree}
:hidden:
:titlesonly:
Home <self>
```

```{toctree}
:hidden:
:titlesonly:
:maxdepth: 6
:caption: Deploy from Snap package
Overview <snap/index.md>
snap/tutorial/index
snap/howto/index
snap/explanation/index
snap/reference/index
```

```{toctree}
:hidden:
:caption: Deploy with Juju
:titlesonly:
:glob:
Overview <charm/index>
charm/tutorial/index
charm/howto/index
charm/explanation/index
charm/reference/index
```

---

````{grid} 1 1 2 2

```{grid-item-card}
:link: snap/
### [Install K8s from a snap ›](snap/index)
^^^
Our tutorials, How To guides and other pages will explain how to install,
 configure and use the Canonical Kubernetes 'k8s' snap.
```

```{grid-item-card}
:link: charm/
### [Deploy K8s using Juju ›](charm/index)
^^^
Our tutorials, How To guides and other pages will explain how to install,
 configure and use the Canonical Kubernetes 'k8s' charm.
```

````
---

## Project and community

Canonical Kubernetes is a member of the Ubuntu family. It's an open source
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
[community]: snap/reference/community
[contribute]: snap/howto/contribute
[roadmap]: snap/reference/roadmap
[overview page]: snap/explanation/about
[architecture documentation]: snap/reference/architecture
