# Security

This page provides an overview of various aspects of security to be considered
when operating a cluster with **{{product}}**. To consider security
properly, this means not just aspects of Kubernetes itself, but also how and
where it is installed and operated.

A lot of important aspects of security therefore lie outside the direct scope
of **{{product}}**, but links for further reading
are provided.

## Security of the snap/executable

As detailed in the [snap documentation], an application installed from a snap
is inherently more secure than a traditionally installed application. Snap-based
applications are installed into a sandboxed, self contained
environment which restricts its ability to interact with the rest of user space.

## Security of the charm

There are several security considerations that must be taken into account when
deploying any charm as outlined in the [Juju security documentation]. With
regards the `k8s` and `k8s-worker` charms, there must be particular care given
to ensuring the principle of least privilege is observed and users only have
access to alter cluster resources they are entitled to. For more information on
creating users, assigning access levels and what access these levels bestow,
please check the following pages of Juju documentation:

- [Juju user types] - describes the different types of users supported by Juju and
their abilities.
- [Working with multiple users] - A how-to guide on sharing control of a cluster
with multiple Juju users.
- [Machine authentication] - describes how SSH keys are stored and used by Juju.

## Security of the OCI images

**{{product}}** relies on OCI standard images published as `rocks` to
deliver the services which run and facilitate the operation of the Kubernetes
cluster. The use of Rockcraft and `rocks` gives Canonical a way to maintain and
patch images to remove vulnerabilities at their source, which is fundamental to
our commitment to a sustainable Long Term Support(LTS) release of Kubernetes
and overcoming the issues of stale images with known vulnerabilities. For more
information on how these images are maintained and published, see the
[Rockcraft documentation][rocks-security].

## Kubernetes Security

The Kubernetes cluster deployed by {{product}} can be secured using
any of the methods and options described by the upstream
[Kubernetes Security Documentation][].

{{product}} enables RBAC (Rules Based Access Control) by default.

## Cloud security

If you are deploying **{{product}}** on public or private cloud
instances, anyone with credentials to the cloud where it is deployed may also
have access to your cluster. Describing the security mechanisms of these clouds
is out of the scope of this documentation, but you may find the following links
useful.

- [Amazon Web Services security][]
- [Google Cloud Platform security][]
- [Metal As A Service(MAAS) hardening][]
- [Microsoft Azure security][]
- [VMware VSphere hardening guides][]

## Security compliance

{{product}} aims to comply with industry security standards by default.
These include the [Centre for Internet Security Kubernetes] (CIS) benchmark and
the Defense Information System Agency (DISA) Security Technical Implementation
Guides (STIG) for [Kubernetes]. {{product}} has applied majority of the
recommended hardening steps in these standards. However, implementing some of
the guidelines would come at the expense of compatibility and/or performance of
the cluster. Therefore, it is expected that cluster administrators follow the
post deployment hardening steps listed in our [hardening guide] and enforce
any of the remaining guidelines according to their needs. Read more about CIS
hardening on our [CIS explanation page].


<!-- LINKS -->
[Juju security documentation]:https://canonical-juju.readthedocs-hosted.com/en/latest/user/explanation/juju-security/
[Machine authentication]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/reference/ssh-key/
[Working with multiple users]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/howto/manage-users
[Juju user types]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/reference/user/
[CIS explanation page]: ./cis
[hardening guide]: ../howto/security/hardening
[Centre for Internet Security Kubernetes]: https://www.cisecurity.org/benchmark/kubernetes
[Kubernetes]: https://www.stigviewer.com/stig/kubernetes/
[Kubernetes Security documentation]: https://kubernetes.io/docs/concepts/security/overview/
[snap documentation]: https://snapcraft.io/docs/security-policies
[rocks-security]: https://documentation.ubuntu.com/rockcraft/en/latest/explanation/rockcraft/
[Amazon Web Services security]: https://aws.amazon.com/security/
[Google Cloud Platform security]:https://cloud.google.com/security
[Metal As A Service(MAAS) hardening]:https://maas.io/docs/how-to-enhance-maas-security
[Microsoft Azure security]:https://docs.microsoft.com/en-us/azure/security/azure-security
[VMware VSphere hardening guides]: https://www.vmware.com/security/hardening-guides.html
[architecture reference page]: ../reference/architecture
