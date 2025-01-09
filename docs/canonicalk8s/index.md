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

src/about.md
Choose an installation method <src/snap/explanation/installation-methods.md>
Deploy from Snap package <src/snap/index.md>
Deploy with Juju <src/charm/index.md>
Deploy with Cluster API <src/capi/index.md>
Community <src/community.md>
Release notes <src/releases.md>

```

````{grid} 1 1 2 2

```{grid-item-card}
:link: src/snap/
### [Install K8s from a snap ›](src/snap/index)
^^^
Our tutorials, How To guides and other pages will explain how to install,
 configure and use the {{product}} 'k8s' snap. This is a great option if you are new to Kubernetes.
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

```{grid-item-card}
:link: about
### [Overview of {{product}} ›](about)
^^^
Find out more about {{product}}, what services are included and get the
answers to some common questions.
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
- Our [release notes][releases]

<!-- IMAGES -->

[logo]: https://assets.ubuntu.com/v1/843c77b6-juju-at-a-glace.svg

<!-- LINKS -->

[Code of Conduct]: https://ubuntu.com/community/ethos/code-of-conduct
[community]: src/snap/reference/community
[contribute]: src/snap/howto/contribute
[releases]: src/snap/reference/releases
[overview page]: about
[architecture documentation]: src/snap/reference/architecture
