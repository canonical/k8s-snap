---
myst:
  html_meta:
    description: Overview of security in Canonical Kubernetes 
---

# Security in {{product}}

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

## Security pages

- [Security explanation]
- [How to report a security issue]

## Updates

Keeping up to date with the latest security updates is an important part of
security maintenance. Read the latest release notes and learn how to upgrade
your cluster.

- [Release notes]
- {ref}`How to upgrade minor version <minor-upgrades>`
- {ref}`How to upgrade patch version <patch-upgrades>`

## Reference material

Our reference material contains technical information that can be used to
understand the security posture of {{product}}, such as which ports are 
exposed, the available security configuration options during bootstrap, 
and much more.

- [Ports and services]
- [Charm configurations]


<!-- LINKS -->
[Ports and services]:/charm/reference/ports-and-services.md
[Release notes]:/releases/charm/index.md
[Security explanation]: /charm/explanation/security.md
[Charm configurations]: /charm/reference/charm-configurations.md
[How to report a security issue]: /charm/howto/report-security-issue.md
