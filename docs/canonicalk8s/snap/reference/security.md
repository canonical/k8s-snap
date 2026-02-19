# Security in {{product}}

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

## Security pages

- [Security explanation]
- [How to report a security issue]
- [How to harden your cluster]

## Authentication

- [How to refresh Kubernetes certificates]
- [How to use intermediate CAs with Vault]
- [Cluster certificates and configuration reference]


## Compliance

- [Assess for CIS compliance]
- [CIS compliance audit]
- [How to install a DISA STIG cluster]
- [DISA STIG compliance audit]
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

- [Ports and services]
- [Configuration files]

<!-- LINKS -->
[Ports and services]:/snap/reference/ports-and-services.md
[Configuration files]:/snap/reference/config-files/index
[How to report a security issue]:/snap/howto/security/report-security-issue.md
[Cluster certificates and configuration reference]: /snap/reference/certificates/
[How to refresh Kubernetes certificates]:/snap/howto/security/refresh-certs.md
[How to use intermediate CAs with Vault]:/snap/howto/security/intermediate-ca.md
[DISA STIG compliance audit]:/snap/reference/disa-stig-audit.md
[How to install a DISA STIG cluster]: /snap/howto/install/disa-stig.md
[CIS compliance audit]: /snap/reference/cis-audit.md
[Assess for CIS compliance]: /snap/howto/security/cis-assessment.md
[How to deploy a cluster with FIPS]: /snap/howto/install/fips.md
[Release notes]:/releases/snap/index.md
[How to upgrade the Canonical Kubernetes snap]:/snap/howto/upgrades.md
[How to install in air-gapped environments]:/snap/howto/install/offline/
[How to harden your cluster]: /snap/howto/security/hardening.md
[Security explanation]: /snap/explanation/security.md
