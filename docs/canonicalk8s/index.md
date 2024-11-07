# {{product}} documentation

{{product}} is a performant, lightweight, secure and
opinionated distribution of **Kubernetes** which includes everything needed to
create and manage a scalable cluster suitable for all use cases.

You can find out more about {{product}} on this [overview page] or
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
Overview <src/snap/index.md>
src/snap/tutorial/index
src/snap/howto/index
src/snap/explanation/index
src/snap/reference/index
```

```{toctree}
:hidden:
:caption: Deploy with Juju
:titlesonly:
:glob:
Overview <src/charm/index>
src/charm/tutorial/index
src/charm/howto/index
src/charm/explanation/index
src/charm/reference/index
```

```{toctree}
:hidden:
:caption: Deploy with Cluster API (WIP)
:titlesonly:
:glob:
Overview <src/capi/index>
src/capi/tutorial/index
src/capi/howto/index
src/capi/explanation/index
src/capi/reference/index
```

---

````{grid} 1 1 2 2

```{grid-item-card}
:link: src/snap/
### [Install K8s from a snap ›](src/snap/index)
^^^
Our tutorials, How To guides and other pages will explain how to install,
 configure and use the {{product}} 'k8s' snap.
```

```{grid-item-card}
:link: src/charm/
### [Deploy K8s using Juju ›](src/charm/index)
^^^
Our tutorials, How To guides and other pages will explain how to install,
 configure and use the {{product}} 'k8s' charm.
```


```{grid-item-card}
:link: src/capi/
### [Deploy K8s using Cluster API ›](src/capi/index)
^^^
Our tutorials, guides and explanation pages will explain how to install,
 configure and use {{product}} through CAPI.
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
[community]: src/snap/reference/community
[contribute]: src/snap/howto/contribute
[roadmap]: src/snap/reference/roadmap
[overview page]: src/snap/explanation/about
[architecture documentation]: src/snap/reference/architecture
