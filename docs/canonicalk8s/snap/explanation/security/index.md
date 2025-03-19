# Security in {{product}}

```{toctree}
:hidden:
Overview <self>
```

This page provides links to the various pages across our documentation that
have security implications on {{product}}.

```{toctree}
:titlesonly:
Overview <security>
cryptography
cis
certificates
```

## Product overview


Architecture: provides the architectural components of Canonical Kubernetes, key to understand the different components to secure in a cluster.

Bootstrap configuration file reference: provides the format of this file by listing all available options and their details, including:
how to setup the default tls secret for ingress
the datastore certificates when an external datastore (like etcd) is used
the certificates to be used for Kubernetes services, front proxy, kube-apiserver, kubelet, kube-proxy, kube-scheduler, service account and admin client


```{toctree}
:titlesonly:
Overview <security>
/snap/reference/architecture
/snap/reference/ports-and-services.md
/snap/reference/config-files/index
```

/snap/howto/security/report-security-issue.md

## Authentication

```{toctree}
:titlesonly:
Certificates explanation <certificates>
/snap/howto/refresh-certs.md
/snap/howto/intermediate-ca.md
Cluster certificates and configuration reference </snap/reference/certificates.md>
```

## Compliance

```{toctree}
:titlesonly:
CIS hardening explanation <cis>
/snap/howto/security/cis-assessment.md
/snap/howto/security/disa-stig-assessment.md
```

## Updates

```{toctree}
:titlesonly:
/snap/reference/releases.md
/snap/howto/upgrades.md
```





Security: specific page about how security is covered in Canonical Kubernetes.

Certificates: explanation page about certificates' role to secure the cluster’s components.

Cluster Certificates and Configuration Directories: provides an overview of certificate authorities (CAs), certificates and configuration directories in use by a Canonical Kubernetes cluster.


Refreshing Kubernetes Certificates: this how-to walks through the steps to refresh the certificates for both control plane and worker nodes in a Canonical Kubernetes cluster.
