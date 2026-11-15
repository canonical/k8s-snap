---
myst:
  html_meta:
    description: "Canonical Kubernetes documentation: a full Kubernetes distribution available as a snap, Juju charm, or via Cluster API."
---
# {{product}} documentation

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

### Getting started

````{domain}
```{slice} Canonical Kubernetes
{doc}`Tutorial <snap/tutorial/getting-started>`
{doc}`What is Canonical Kubernetes? <about>`
{doc}`Snap, charm or CAPI? <snap/explanation/installation-methods>`
```
````

### Deployment

````{domain}
```{slice} Snap
{doc}`k8s snap <snap/howto/install/snap>`
{doc}`Custom bootstrap config <snap/howto/install/custom-bootstrap-config>` slice
{doc}`Add a worker <snap/howto/install/custom-worker>` slice
{doc}`Development environments <snap/howto/install/dev-env>`
{doc}`Multipass <snap/howto/install/multipass>`
{doc}`LXD <snap/howto/install/lxd>` slice
```

```{slice} Charm
{doc}`k8s charms <charm/howto/install/charm>`
{doc}`Custom bootstrap config <charm/howto/install/install-custom>` slice
{doc}`Add a worker <charm/howto/install/custom-workers>` slice
{doc}`LXD <charm/howto/install/install-lxd>` slice
{doc}`Terraform <charm/howto/install/install-terraform>`
```

```{slice} CAPI
{doc}`k8s CAPI <capi/howto/provision>`
{doc}`Custom bootstrap config <capi/howto/custom-bootstrap-config>` slice
{doc}`Custom Kubernetes version <capi/howto/custom-ck8s>`
```
````

### Architecture and core features

````{domain}
```{slice} Architecture
{doc}`Overview <snap/explanation/architecture>` slice
{doc}`Cluster API and Canonical Kubernetes <capi/explanation/capi-ck8s>`
```

```{slice} Networking
{doc}`Overview <snap/explanation/networking>` slice
{doc}`DNS <snap/howto/networking/default-dns>`
{doc}`Network <snap/howto/networking/default-network>`
{doc}`Load balancer <snap/howto/networking/default-loadbalancer>`
{doc}`Ingress <snap/howto/networking/default-ingress>`
{doc}`Gateway <snap/howto/networking/default-gateway>`
{doc}`Dual stack <snap/howto/networking/dualstack>`
{doc}`IPv6 only <snap/howto/networking/ipv6>`
{doc}`Multi-peer BGP <snap/howto/networking/multi-peer-bgp>`
{doc}`Ports and services <snap/reference/ports-and-services>`
```

```{slice} Proxied networking
{doc}`Configure proxy <snap/howto/networking/proxy>`
{doc}`Proxy environment variables <snap/reference/proxy>`
```

```{slice} Storage
{doc}`Use default storage <snap/howto/storage/storage>`
{doc}`etcd <snap/reference/etcd>` slice
{doc}`Dqlite <snap/reference/dqlite>`
{doc}`Use an external datastore <snap/howto/external-datastore>`
{doc}`Backup and restore <snap/howto/backup-restore>`
```

```{slice} Monitoring
{doc}`Observability <snap/howto/observability>`
```

```{slice} Scaling
{doc}`High availability <snap/explanation/high-availability>`
{doc}`Clustering <snap/explanation/clustering>`
{doc}`Node roles <snap/explanation/roles>`
```

```{slice} Configuration
{doc}`Configuration files <snap/reference/config-files/index>`
{doc}`Annotations <snap/reference/annotations>`
{doc}`Commands <snap/reference/commands>`
{doc}`Availability zones <charm/reference/az>`
{doc}`Actions <charm/reference/actions>`
{doc}`Provider configuration <capi/reference/configs>`
```
````

### External integrations

````{domain}
```{slice} Charm integrations
{doc}`OpenStack <charm/howto/openstack>`
{doc}`etcd <charm/howto/etcd>` slice
{doc}`Ceph-CSI <charm/howto/ceph-csi>`
{doc}`COS Lite <charm/howto/cos-lite>`
```
````

### Security

````{domain}
```{slice} Security
{doc}`Overview <snap/explanation/security>` slice
{doc}`Cluster hardening <snap/howto/security/hardening>`
{doc}`Refresh certificates <snap/howto/security/refresh-certs>`
{doc}`Air-gapped deployments <snap/howto/install/offline>`
{doc}`Configure firewall <snap/howto/networking/ufw>`
{doc}`Cluster certificates <snap/reference/certificates>`
```

```{slice} Compliance
{doc}`DISA STIG <snap/howto/install/disa-stig>`
{doc}`CIS <snap/howto/security/cis-assessment>`
{doc}`FIPS <snap/howto/install/fips>`
```
````

### Lifecycle management

````{domain}
```{slice} Installation
{doc}`Choose a channel <snap/explanation/channels>`
```

```{slice} Troubleshooting
{doc}`Troubleshoot your cluster <snap/howto/troubleshooting>`
{doc}`Get support <snap/howto/support>`
```

```{slice} Disaster recovery
{doc}`Recover after quorum loss <snap/howto/restore-quorum>`
{doc}`Inspection reports <snap/reference/inspection-reports>`
```

```{slice} Upgrades
{doc}`Overview <snap/explanation/upgrade>` slice
{doc}`Manage upgrades <snap/howto/upgrades>`
{doc}`In place upgrades <capi/explanation/in-place-upgrades>`
{doc}`Validate your cluster <charm/howto/validate>`
```

```{slice} Images
{doc}`Manage images <snap/howto/image-management>`
```

```{slice} Migration
{doc}`Migrate CAPI cluster <capi/howto/migrate-management>`
```
````

---

## Project and community

{{product}} is a member of the Ubuntu family. It's an open source
project which welcomes community involvement, contributions, suggestions, fixes
and constructive feedback.

### Get involved

- [Canonical Kubernetes Slack]
- [Canonical Kubernetes Discourse]
- [Community]
- [How to contribute]

### Releases 

- [Release notes][releases]

### Governance and policies

- [Code of Conduct]

### Commercial support

Thinking about using {{product}} for your next project? [Get in touch!]

<!-- IMAGES -->

[logo]: https://assets.ubuntu.com/v1/843c77b6-juju-at-a-glace.svg

<!-- LINKS -->

[Code of Conduct]: https://ubuntu.com/community/ethos/code-of-conduct
[community]: /community
[How to contribute]: /snap/howto/contribute
[releases]: /releases/index
[Canonical Kubernetes Slack]: https://kubernetes.slack.com/archives/CG1V2CAMB
[Canonical Kubernetes Discourse]: https://discourse.ubuntu.com/c/kubernetes/180
[Get in touch!]: https://ubuntu.com/kubernetes/contact-us
