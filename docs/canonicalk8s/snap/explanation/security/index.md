# Security in {{product}}

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

## Security pages

Visit our dedicated security explanation pages to learn more in depth about
security in {{product}}.

```{toctree}
:titlesonly:
Security overview <security>
certificates
cis
```
<!-- Add back in once we have this page complete Cryptography <cryptography> -->

We also provide how-to guides on reporting a security vulnerability and
hardening your cluster.

- [How to report a security issue]
- [How to harden your cluster]

## Authentication

The use of certificates to authenticate and secure communication within the
cluster is an important part of Kubernetes' security infrastructure. Read our
explanation page detailing how they are implemented in {{product}} as well as
how-to guides on managing your cluster certificates.

- [Certificates explanation]
- [How to refresh Kubernetes certificates]
- [How to use intermediate CAs with Vault]
- [Cluster certificates and configuration reference]


## Compliance

Read our explanation page on what CIS hardening means in terms of {{product}} or
follow our how-to guides to assess your cluster for compliance.

- [CIS hardening explanation]
- [How to assess for CIS compliance]
- [How to assess for DISA STIG compliance]

## Updates

Keeping up to date with the latest security updates is an important part of
security maintenance. Read the latest release notes and learn how to upgrade
your cluster.

- [Release notes]
- [How to upgrade the Canonical Kubernetes snap]

## Air-gapped deployment

If you would like to install {{product}} in an air-gapped environment, we have
the following guide to help:

- [How to install in air-gapped environments]

## Reference material

Our reference material contains technical information that can be used to
understand the security posture of {{product}} such as what ports are exposed,
what are the different security configuration options available during bootstrap
and much more.

- [Architecture]
- [Ports and services]
- [Configuration files]

<!-- LINKS -->
[Architecture]:/snap/reference/architecture
[Ports and services]:/snap/reference/ports-and-services.md
[Configuration files]:/snap/reference/config-files/index
[How to report a security issue]:/snap/howto/security/report-security-issue.md
[Cluster certificates and configuration reference]: /snap/reference/certificates/
[How to refresh Kubernetes certificates]:/snap/howto/refresh-certs.md
[How to use intermediate CAs with Vault]:/snap/howto/intermediate-ca.md
[How to assess for DISA STIG compliance]:/snap/howto/security/disa-stig-assessment.md
[How to assess for CIS compliance]: /snap/howto/security/cis-assessment.md
[Release notes]:/snap/reference/releases.md
[How to upgrade the Canonical Kubernetes snap]:/snap/howto/upgrades.md
[Certificates explanation]: certificates
[CIS hardening explanation]: cis
[How to install in air-gapped environments]:/snap/howto/install/offline/
[How to harden your cluster]: /snap/howto/security/hardening.md
