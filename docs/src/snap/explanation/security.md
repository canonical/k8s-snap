# Security

This page provides an overview of various aspects of security to be considered
when operating a cluster with **{{product}}**. To consider security
properly, this means not just aspects of Kubernetes itself, but also how and
where it is installed and operated.

A lot of important aspects of security therefore lie outside the direct scope
of **{{product}}**, but links for further reading
are provided.

## Security of the snap/executable

As detailed in the [snap documentation][], an application installed from a snap
is inherently more secure than a traditionally installed application.
Snap-based applications are installed into a sandboxed, self contained
environment which restricts its ability to interact with the rest of user
space.

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

- Amazon Web Services <https://aws.amazon.com/security/>
- Google Cloud Platform <https://cloud.google.com/security/>
- Metal As A Service(MAAS) <https://maas.io/docs/snap/3.0/ui/hardening-your-maas-installation>
- Microsoft Azure <https://docs.microsoft.com/en-us/azure/security/azure-security>
- VMWare VSphere <https://www.vmware.com/security/hardening-guides.html>

## Security Compliance

As with previously released Kubernetes software from Canonical, we aim to
satisfy the needs of various security compliance standards. This is a process
that will take some time however. Please watch out for future announcements and
check the [roadmap][] for current areas of work.

<!-- LINKS -->

[Kubernetes Security documentation]: https://kubernetes.io/docs/concepts/security/overview/
[snap documentation]: https://snapcraft.io/docs/security-sandboxing
[rocks-security]: https://canonical-rockcraft.readthedocs-hosted.com/en/latest/explanation/rockcraft/
[roadmap]: /snap/reference/roadmap
