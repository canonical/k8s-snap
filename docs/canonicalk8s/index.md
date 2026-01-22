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
your needs. If you would like to jump straight in, head to the
[snap getting started tutorial!](/snap/tutorial/getting-started.md)

![Illustration depicting working on components and clouds][logo]

---

## In this documentation
<!-- markdownlint-disable -->
{{product}} can be deployed and managed as a standalone snap, as a charm as part of a
Juju cluster or with Cluster API. Find out more about which {{product}}
deployment method is best for your
project's needs with
**[choosing a {{product}} installation method.](/snap/explanation/installation-methods.md)**
<!-- markdownlint-restore -->

```{toctree}
:hidden:
:titlesonly:
Canonical Kubernetes documentation <self>
```

```{toctree}
:hidden:
:titlesonly:
:maxdepth: 6

about.md
Deploy from Snap package </snap/index.md>
Deploy with Juju </charm/index.md>
Deploy with Cluster API </capi/index.md>
Community </community.md>
Release notes </releases/index.md>
```

````{grid} 3

```{grid-item-card}
:link: snap/
### [Canonical Kubernetes snap ›](/snap/index)

The `k8s` snap is a self-contained, secure and dependency-free Linux app package used to deploy and manage a {{product}} cluster. If you are new to Kubernetes, start here.
```

```{grid-item-card}
:link: charm/
### [Canonical Kubernetes charms ›](/charm/index)

The `k8s` charms take care of installing, configuring and managing {{product}} on cloud instances managed by Juju.
```

```{grid-item-card}
:link: capi/
### [Canonical Kubernetes and Cluster API ›](/capi/index)

Using Cluster API's declarative tooling, deploy and manage multiple {{product}} clusters.
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
[community]: /snap/reference/community
[contribute]: /snap/howto/contribute
[releases]: /releases/index
