# Security in {{product}}

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

## Security pages

- [Security explanation]
- [How to report a security issue]
- [How to harden your cluster]
- [Cryptography in Canonical Kubernetes explanation]

## Authentication

- [Certificates explanation]
- [How to refresh Kubernetes certificates]
- [How to use intermediate CAs with Vault]
- [Cluster certificates and configuration reference]


## Compliance

- [CIS hardening explanation]
- [Audit for CIS compliance]
- [How to install a DISA STIG cluster]
- [DISA STIG for Kubernetes explanation]
- [Audit for DISA STIG compliance]
- [How to deploy a cluster with FIPS]

## Updates

Keeping up to date with the latest security updates is an important part of
security maintenance. Read the latest release notes and learn how to upgrade
your cluster.

- [Release notes]
- [How to upgrade the Canonical Kubernetes snap]

## Air-gapped deployment

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
[How to refresh Kubernetes certificates]:/snap/howto/security/refresh-certs.md
[How to use intermediate CAs with Vault]:/snap/howto/security/intermediate-ca.md
[Audit for DISA STIG compliance]:/snap/reference/disa-stig-audit.md
[DISA STIG for Kubernetes explanation]: /snap/explanation/security/#kubernetes-disa-stig
[How to install a DISA STIG cluster]: TODO
[Audit for CIS compliance]: /snap/reference/cis-audit.md
[How to deploy a cluster with FIPS]: TODO
[Release notes]:/snap/reference/releases.md
[How to upgrade the Canonical Kubernetes snap]:/snap/howto/upgrades.md
[Certificates explanation]: /snap/explanation/security/#certificates
[CIS hardening explanation]: /snap/explanation/security/#cis-hardening
[How to install in air-gapped environments]:/snap/howto/install/offline/
[How to harden your cluster]: /snap/howto/security/hardening.md
[Security explanation]: /snap/explanation/security.md
[CIS hardening]: /snap/explanation/security/cis
[Cryptography in Canonical Kubernetes explanation]: /snap/explanation/security/#cryptography
