<!-- Start -->
# Security in {{product}}

This page provides an insight into the various aspects of security to be
considered when operating a cluster with {{product}}. To consider security
properly, this means not just aspects of Kubernetes itself, but also how and
where it is installed and operated.

A lot of important aspects of security therefore lie outside the direct scope
of {{product}}, but links for further reading are provided.

## Security of the snap/executable

Keeping up to date with the latest security patches is one of the most
effective ways to keep your cluster secure. Deploying {{product}} as a snap
allows our users to automatically consume the latest security patches with snap
refreshes taking place several times a day. The `k8s` snap is deployed with
`classic`confinement meaning that the snap has access to system resources in
order to be able to deploy the cluster successfully. See the
[Snapcraft documentation] for more
information on confinement levels and security in snaps.

<!-- First charm end here -->

## Security of the OCI images

{{product}} relies on OCI standard images published as `rocks` to
deliver the services which run and facilitate the operation of the Kubernetes
cluster. The use of Rockcraft and `rocks` gives Canonical a way to maintain and
patch images to remove vulnerabilities at their source, which is fundamental to
our commitment to a sustainable Long Term Support(LTS) release of Kubernetes
and overcoming the issues of stale images with known vulnerabilities. For more
information on how these images are maintained and published, see the
[Rockcraft documentation][rocks-security].

## Authentication and authorization

{{product}} leverages upstream Kubernetes security primitives for both
authentication (identity verification) and authorization (determining
permissions).

### Authentication

Authentication in {{product}} ensures that users and components are who they
claim to be. The following methods are included by default:

#### Client certificates

- The Kubernetes API server is configured to trust specific client certificates.
- Certificates are issued to admin users during cluster creation.
- These can be seen on the snap by running `k8s config` on a control plane 
node.

#### Service accounts

- Every pod in Kubernetes is automatically assigned a service account, unless
  specified otherwise.
- Service account tokens are mounted into pods, enabling them to authenticate
  with the API server securely, unless specified otherwise.
- A policy engine (if added) may restrict auto-mounting the default service
  account token into Pods if used.
- These are managed in the namespace where the pod is deployed.

The Kubernetes API server can be configured to accept OpenID Connect (OIDC)
tokens for authentication from external identity providers.

In {{product}}, anonymous API access is disabled by default.

For more details, see [Kubernetes Authentication]
upstream docs.

### Authorization

After authentication, Kubernetes checks whether the user or service account is
authorized to perform a requested action. In {{product}}, this is done through
Role-Based Access Control.

#### Role-Based Access Control (RBAC)

RBAC Authorization in {{product}} is done through the following types of
Kubernetes objects:

- **Roles** define a set of _allow_ permissions within a namespace.
- **ClusterRoles** define cluster-wide _allow_ permissions.
- **RoleBindings** and **ClusterRoleBindings** assign these roles to users or
  service accounts.

Example use cases include granting users or service accounts read-only access to
logs in a namespace or granting read-only access to nodes, services, and pods to
a monitoring operator.

Kubernetes defines a set of `ClusterRoles` and `ClusterRoleBindings` that apply
to all users and service accounts, including the `default` service accounts
mounted in pods:

- `system:basic-user`: Allows a user read-only access to basic information about
  themselves.
- `system:discovery`: Allows read-only access to API discovery endpoints needed
  to discover and negotiate an API level.
- `system:public-info-viewer`: Allows read-only access to non-sensitive
  information about the cluster.

For more details, see the upstream [Kubernetes RBAC] and
[Default roles and role bindings] documentation.

#### Admission controllers

After a request has been authenticated and authorized, admission controllers
are used to **validate** and / or **mutate** requests to the Kubernetes API that
create, modify, or delete resources in Kubernetes. Admission controllers cannot
block **get**, **list**, or **watch** requests.

Kubernetes has a default list of admission controllers, which can be expanded
or contracted. Keep in mind that several important features of Kubernetes
require admission controller to be enabled in order to properly support them.
As a result, a Kubernetes API server that is not properly configured with the
right set of admission controllers is an incomplete server and will not support
all the features you expect.

In addition to the default list of admission controllers, {{product}} enables
the `NodeRestriction` admission controller. This controller limits the `Node`
and `Pod` objects a kubelet can modify, allowing them to modify only their
own `Node` object and `Pod` objects bound to their node.

For more details, see [Kubernetes Admission Control] upstream docs.

### Extending authentication and authorization

For advanced authentication and authorization, users may want to configure:

- OpenID Connect (OIDC) for integrating with external identity providers
  (e.g.: LDAP, Google, Azure AD).
- Webhook token authentication for custom authentication.
- Custom RBAC roles.
- Pod Security Admission Controller, which can enforce the Pod Security
  Standards.

For further information on how to extend authentication and authorization in
your cluster, see:

- [Kubernetes authentication strategies]
- [Webhook token authentication]
- [Managing RBAC roles]
- [Pod security admission]

## Certificates

Certificates are a crucial part of Kubernetes' security infrastructure, serving
to authenticate and secure communication within the cluster. They play a key
role in ensuring that communication between various components (such as the
API server, kubelet, and the datastore) is both encrypted and restricted to
authorized components only.

In Kubernetes, [X.509] certificates are primarily used for
[Transport Layer Security] (TLS), securing the communication between the
cluster's components.

### What is a certificate refresh?

A certificate refresh in {{product}} refers to the process of renewing or
rotating certificates before they expire. Kubernetes certificates have
a specific validity period, after which they expire and are no longer
considered valid. Expired certificates lead to failures in communication
between cluster components, potentially disrupting the functionality of the
entire cluster.

### Importance of certificate refreshes

Regularly refreshing certificates in Kubernetes is essential for maintaining
the reliability of the cluster. Here are some reasons why certificate refreshes
are important:

- **Maintaining cluster security** - Certificates are crucial for securing
Kubernetes clusters, ensuring encrypted and authenticated communication between
components. If a certificate expires
and isn't promptly renewed, it can leave the cluster vulnerable to security
risks, potentially exposing it to unauthorized access. Regular certificate
refreshes prevent this by ensuring only valid certificates are used,
maintaining the cluster's security.
- **Preventing downtime** - Kubernetes relies on certificates for internal
communication between critical
components, such as the kubelet, API server, and datastore. Expired
certificates can hinder this communication, leading to potential downtime and
workload disruptions. Proactively refreshing certificates before they expire
helps maintain uninterrupted cluster operations.
- **Security compliance** - Security standards, such as [CIS][], often require
the regular rotation of credentials, including certificates. Periodically
renewing Kubernetes certificates ensures that the cluster meets security
standards and compliance requirements.

### Performing certificate refreshes

If you would like to refresh the certificates of your {{product}} cluster see
our [refreshing certificates how-to guide].

## Cryptography

{{product}} uses industry-standard cryptographic
algorithms to ensure authentication between components, secure data transfer,
and data encryption at rest.

### TLS certificates

All communications between core components, such as the API server and
kubelets, are encrypted with TLS 1.3 (Transport Layer Security), providing
robust protection for sensitive data in transit. By default, {{product}} uses
self-signed certificates, but users are able to use an intermediate CA or
provide their own certificates instead.

### Encryption at rest

{{product}} uses AES-256-GCM (Advanced Encryption Standard - Galois/Counter
Mode) to encrypt cluster data at rest.

### Digital signatures

To securely authenticate API clients, Canonical Kubernetes uses X.509
certificates with support for both RSA-2048 and ECDSA (Elliptic Curve Digital
Signature Algorithm). This ensures strong, standards-based authentication with
options suited for both general-purpose and resource-constrained environments.

### Configure cryptography

Canonical Kubernetes provides various cryptographic tools that users can
leverage to implement security controls for their workloads:

* **Kubernetes Secrets Encryption**: Users are empowered to configure encryption
providers for Secrets at rest, with AES-GCM as the preferred encryption
algorithm for data confidentiality.
* **Kubernetes API Authentication**: For secure API access, users can configure
X.509 certificates, allowing them to implement secure, certificate-based
authentication with support for RSA 2048 and ECDSA keys.
* **Service Mesh Encryption (Optional)**: When deploying service meshes like
Istio with Canonical Kubernetes, users can enable mutual TLS (mTLS) to protect
inter-service communications, ensuring data privacy and authenticity in
multi-service environments. Supported Algorithms for mTLS:
  * RSA-2048 or ECDSA: These options enable secure certificate-based
  authentication between services.
  * AES-GCM: Used for encrypted service-to-service communication.

### Third party cryptographic packages and libraries

Canonical Kubernetes depends on a suite of cryptographic libraries and packages
to implement its security functions:

* **OpenSSL**: Canonical Kubernetes utilizes OpenSSL for a broad range of
cryptographic operations, including TLS, certificate management, and secure key
exchange. OpenSSL’s extensive cryptographic functionality and secure algorithms
provide a reliable foundation for TLS and encryption operations. Source: Ubuntu
Archive (Package: [openssl])

* **Linux Kernel Cryptographic Modules**: For network security and cryptographic
operations at the kernel level, Canonical Kubernetes leverages cryptographic
modules in the Linux kernel, such as IPsec for secure network communications.
Source: Ubuntu Kernel (Package: [linux-generic])

* **Go Cryptography Library**: Since Kubernetes is written in Go, Canonical
Kubernetes relies on the Go standard library's cryptographic functions,
implementing secure algorithms such as RSA, ECDSA, and AES, which are necessary
for the secure operation of Kubernetes components. Source: Go Standard Library
([x509], [rsa], [sha256], [tls])

### Recommended crypto usage and settings

{{product}} ships with a secure-by-default security posture, so users can rest
assured that the default configuration is appropriate for most use-cases. If
your security needs are not met by the default configuration, we recommend you
[deploy an intermediate CA fine-tuned to your liking][intermediate-ca]. If you
require the use of FIPS compliant cryptographic libraries, continue reading
about security compliance in {{product}} snap.

## Security compliance

{{product}} snap aims to comply with industry security standards by default
and has applied majority of the recommended hardening steps for standards such
as the CIS Kubernetes benchmark and Defense Information System Agency (DISA)
Security Technical Implementation Guides (STIG) for Kubernetes.
However, implementing some of the guidelines would come at the expense of
compatibility and/or performance of the cluster. Therefore, it is expected that
cluster administrators follow the post deployment hardening steps listed in our
[hardening guide] and enforce any of the remaining guidelines according to their
needs.

### CIS hardening

CIS hardening refers to the process of implementing security configurations that
align with the benchmarks set forth by the [Center for Internet Security] (CIS).
These [benchmarks] are a set of best practices and guidelines designed to secure
various software and hardware systems, including Kubernetes clusters. The
primary goal of CIS hardening is to reduce the attack surface and enhance the
overall security posture of an environment by enforcing configurations that are
known to protect against common vulnerabilities and threats. Kubernetes, by its
nature, is a complex system with many components interacting
in a distributed environment. This complexity can introduce numerous security
risks if not properly managed such as unauthorized access, data breaches and
service disruption.

If you would like to apply CIS hardening to your {{product}} snap see our
[hardening guide] and follow our [CIS assessment] guide to assess your snap
deployment for compliance.

### FIPS compliance

The [Federal Information Processing Standard] (FIPS) 140-3 is a U.S. government
standard for cryptographic modules. In order to comply with FIPS standards,
each cryptographic module must meet specific security requirements and must
undergo testing and validation by the U.S. National Institute of Standards
and Technology ([NIST]). All of our components including the built-in features
such as networking or load-balancer can be configured in snap deployments to
use host systems FIPS compliant libraries instead of the non-compliant internal
go cryptographic
modules. When building workloads on top of {{ product }}, it is essential that
organizations build these in a FIPS compliant manner to comply with the FIPS
security requirements. In addition, FIPS 140-3 has additional requirements to
the system and hardware that have to be met in order to be fully FIPS compliant.

If you would like to enable FIPS in your Kubernetes cluster see our
[Canonical Kubernetes snap FIPS installation] guide.

### Kubernetes DISA STIG

Security Technical Implementation Guides (STIGs) are developed by the Defense
Information System Agency (DISA) and are comprehensive frameworks of security
requirements designed to protect U.S. Department of Defense (DoD) systems and
networks from cybersecurity threats.
The [Kubernetes STIG] contains guidelines on how to check and remediate various
potential security concerns for a Kubernetes deployment, both on the host and
within the cluster itself.

The [deploy Canonical Kubernetes snap with DISA STIG hardening] guide provides
configuration files to harden your cluster and host set up in accordance with
DISA STIG Kubernetes guidelines. {{product}} snap aligns with many DISA STIG
compliance recommendations by default. However, additional hardening steps are
required to fully meet the standard.

## Cloud security

If you are deploying {{product}} on public or private cloud
instances, anyone with credentials to the cloud where it is deployed may also
have access to your cluster. Describing the security mechanisms of these clouds
is out of the scope of this documentation, but you may find the following links
useful.

- [Amazon Web Services security][]
- [Google Cloud Platform security][]
- [Metal As A Service(MAAS) hardening][]
- [Microsoft Azure security][]
- [VMware VSphere hardening guides][]

<!-- LINKS -->
[intermediate-ca]: /snap/howto/security/intermediate-ca.md
[openssl]: https://packages.ubuntu.com/jammy/openssl
[linux-generic]: https://packages.ubuntu.com/jammy/linux-generic
[x509]: https://pkg.go.dev/crypto/x509
[rsa]: https://pkg.go.dev/crypto/rsa
[sha256]: https://pkg.go.dev/crypto/sha256
[tls]: https://pkg.go.dev/crypto/tls
[Vault]: https://www.hashicorp.com/en/products/vault
[benchmarks]: https://www.cisecurity.org/cis-benchmarks
[Center for Internet Security]: https://www.cisecurity.org/
[CIS assessment]: /snap/howto/security/cis-assessment.md
[hardening guide]: /snap/howto/security/hardening
[Kubernetes STIG]: https://www.stigviewer.com/stig/kubernetes/
[Snapcraft documentation]: https://snapcraft.io/docs/security-policies
[rocks-security]: https://documentation.ubuntu.com/rockcraft/en/latest/explanation/rockcraft/
[Amazon Web Services security]: https://aws.amazon.com/security/
[Google Cloud Platform security]:https://cloud.google.com/security
[Metal As A Service(MAAS) hardening]:https://maas.io/docs/how-to-enhance-maas-security
[Microsoft Azure security]:https://docs.microsoft.com/en-us/azure/security/azure-security
[VMware VSphere hardening guides]: https://www.vmware.com/security/hardening-guides.html
[CIS]: https://www.cisecurity.org/controls
[Transport Layer Security]: https://datatracker.ietf.org/doc/html/rfc8446
[X.509]: https://datatracker.ietf.org/doc/html/rfc5280
[refreshing certificates how-to guide]: /snap/howto/security/refresh-certs.md
[Federal Information Processing Standard]: https://csrc.nist.gov/pubs/fips/140-3/final
[NIST]: https://www.nist.gov/


[Canonical Kubernetes snap FIPS installation]: /snap/howto/install/fips.md
[deploy Canonical Kubernetes snap with DISA STIG hardening]: /snap/howto/install/disa-stig.md
[Kubernetes Authentication]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/
[Kubernetes RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[Default roles and role bindings]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#default-roles-and-role-bindings
[Kubernetes Admission Control]: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/
[Kubernetes authentication strategies]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies
[Webhook token authentication]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
[Managing RBAC roles]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles
[Pod security admission]: https://kubernetes.io/docs/concepts/security/pod-security-admission/

