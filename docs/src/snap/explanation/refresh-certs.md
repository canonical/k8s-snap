# Certificates

## What Are Certificates in Kubernetes?

Certificates are a crucial part of Kubernetes' security infrastructure, serving
to authenticate and secure communication within the cluster. They play a key
role in ensuring that communication between various components (such as the
API server, kubelets, and the datastore) is both encrypted and restricted to
authorized components only.

In Kubernetes, [X.509][] certificates are primarily used for
[Transport Layer Security][] (TLS), securing the communication between the
cluster's components.

## What Is a Certificate Refresh?

A certificate refresh in {{product}} refers to the process of renewing or
rotating certificates before they expire. Kubernetes certificates have
a specific validity period, after which they expire and are no longer
considered valid. Expired certificates lead to failures in communication
between cluster components, potentially disrupting the entire cluster
functionality.

Regular certificate refreshes ensure that the certificates in use are valid and
secure, preventing unauthorized access or communication failures.

## Importance of Certificate Refreshes

### Maintaining Cluster Security

Certificates are crucial for securing Kubernetes clusters, ensuring encrypted
and authenticated communication between components. If a certificate expires
and isn't promptly renewed, it can leave the cluster vulnerable to security
risks, potentially exposing it to unauthorized access. Regular certificate
refreshes prevent this by ensuring only valid certificates are used,
maintaining the cluster's security.

### Preventing Downtime

Kubernetes relies on certificates for internal communication between critical
components, such as the kubelet, API server, and datastore. Expired
certificates can hinder this communication, leading to potential downtime and
workload disruptions. Proactively refreshing certificates before they expire
helps maintain uninterrupted cluster operations.

### Security Compliance

Security standards, such as [CIS][], often require the regular rotation of
credentials, including certificates. Periodically renewing Kubernetes
certificates ensures that the cluster meets security standards and compliance
requirements.

<!-- LINKS -->

[CIS]: https://www.cisecurity.org/controls
[Transport Layer Security]: https://datatracker.ietf.org/doc/html/rfc8446
[X.509]: https://datatracker.ietf.org/doc/html/rfc5280
