# {{product}} documentation

{{product}} is a performant, lightweight, secure and
opinionated distribution of **Kubernetes** which includes everything needed to
create and manage a scalable cluster suitable for all use cases.

{{product}} builds upon upstream Kubernetes by providing all the extra services
such as a container runtime, a CNI, DNS services, an ingress gateway and more
that are necessary to have a fully functioning cluster all in one convenient
location - a snap!

Staying up-to-date with upstream Kubernetes security
patches and updates with {{product}} is a seamless experience, freeing up time
for application
development and innovation without having to worry about the infrastructure.

Whether you are deploying a small cluster to get accustomed to Kubernetes or a
huge enterprise level deployment across the globe, {{product}} can cater to
your needs. If would like to jump straight in, head to the
[snap getting started tutorial!](src/snap/tutorial/getting-started.md)

![Illustration depicting working on components and clouds][logo]

---

## In this documentation
<!-- markdownlint-disable -->
{{product}} can be deployed as a standalone snap, as a charm as part of a
Juju cluster or with Cluster API. Find out more about which {{product}}
installation method is best for your
project's needs with
**[choosing a {{product}} installation method.](src/snap/explanation/installation-methods.md)**
<!-- markdownlint-restore -->

```{toctree}
:hidden:
:titlesonly:
Canonical Kubernetes documentation <self>
```

---

```{toctree}
:hidden:
:titlesonly:
:maxdepth: 6

src/about.md
Deploy from Snap package <src/snap/index.md>
Deploy with Juju <src/charm/index.md>
Deploy with Cluster API <src/capi/index.md>
Community <src/community.md>
Release notes <src/releases.md>

```

````{grid} 1 1 1 1

```{grid-item-card}
:link: src/snap/
### [Install with a snap ›](src/snap/index)

Our tutorials, how-to guides and other pages will explain how to install,
 configure and use the {{product}} 'k8s' snap. If you are new to Kubernetes, start here.
```

```{grid-item-card}
:link: src/charm/
### [Deploy with Juju ›](src/charm/index)

Our tutorials, how-to guides and other pages will explain how to install,
 configure and use the {{product}} 'k8s' charm.
```


```{grid-item-card}
:link: src/capi/
### [Deploy with Cluster API ›](src/capi/index)

Our tutorials, how-to guides and other pages will explain how to install,
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
- Our [release notes][releases]

<!-- IMAGES -->

[logo]: https://assets.ubuntu.com/v1/843c77b6-juju-at-a-glace.svg

<!-- LINKS -->

[Code of Conduct]: https://ubuntu.com/community/ethos/code-of-conduct
[community]: src/snap/reference/community
[contribute]: src/snap/howto/contribute
[releases]: src/snap/reference/releases
