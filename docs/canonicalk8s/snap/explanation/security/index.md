# Security in {{product}}
<!--
```{toctree}
:hidden:
Overview <self>
``` -->

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

<!-- ```{toctree}
:titlesonly:
Overview <security>
cryptography
cis
certificates
``` -->
## Security pages

Visit our dedicated security explanation pages to learn more in depth about
security in {{product}}.

```{toctree}
:titlesonly:
Security overview <security>
Cryptography <cryptography>
certificates
cis
```

We also provide a how-to guide on reporting a security vulnerability.

- [How to report a security issue]

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

<!-- Architecture: provides the architectural components of Canonical Kubernetes, key to understand the different components to secure in a cluster.

Bootstrap configuration file reference: provides the format of this file by listing all available options and their details, including:
how to setup the default tls secret for ingress
the datastore certificates when an external datastore (like etcd) is used
the certificates to be used for Kubernetes services, front proxy, kube-apiserver, kubelet, kube-proxy, kube-scheduler, service account and admin client -->


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
