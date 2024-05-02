# Canonical Kubernetes charm documentation

The Canonical Kubernetes charm, `k8s`, is an operator: software which wraps an
application and contains all of the instructions necessary for deploying,
configuring, scaling, integrating the application on any cloud supported by
[Juju][].

The `k8s` charm takes care of installing and configuring the [k8s snap
package][] on cloud instances managed by Juju. Operating Kubernetes through
this charm makes it significantly easier to manage at scale, on remote cloud
instances and also to integrate other operators to enhance or customise your
Kubernetes deployment. You can find out more about Canonical Kubernetes on this
[overview page][] or see a more detailed explanation in our [architecture
documentation][arch].

![Illustration depicting working on components and clouds][logo]

## In this documentation

````{grid} 1 1 2 2

```{grid-item-card} [Tutorial](tutorial/index)

**Start here!** A hands-on introduction to Canonical K8s for new users
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
[community]: /charm/reference/community
[contribute]: /snap/howto/contribute
[roadmap]: /snap/reference/roadmap
[overview page]: /charm/explanation/about
[arch]: /charm/reference/architecture
[Juju]: https://juju.is
[k8s snap package]: /snap/index