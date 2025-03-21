# Cryptography in {{product}}

<!-- Guidance:
Short description of how cryptographic technology is used in K8s / what
will be covered in this doc -->

<!-- SSDLC:
Canonical Kubernetes employs industry-standard cryptographic protocols and
algorithms to ensure data confidentiality, integrity, and authenticity within
the Kubernetes environment. Primary uses of cryptography include:
- Securing API Communications: All communications between core components, such as
the API server and kubelets, are encrypted with TLS (Transport Layer Security),
providing robust protection for sensitive data in transit.
- Encrypting Secrets and Sensitive Data: Canonical Kubernetes supports AES-GCM
(Advanced Encryption Standard - Galois/Counter Mode) for encrypting Secrets at
rest. This is particularly important for protecting sensitive configurations and
credentials stored within the cluster.
- Digital Signatures for API Authentication: To authenticate API clients securely,
Canonical Kubernetes supports X.509 certificates, leveraging both RSA and ECDSA
(Elliptic Curve Digital Signature Algorithm) for digital signatures, ensuring
robust authentication mechanisms. -->

## Cryptographic technology used by Canonical Kubernetes
<!-- Guidance:
Cryptographic technology used by the product
In this section list/describe the places where cryptographic technology is used
internally by the product itself. In this section should list the algorithms and
key lengths used (where applicable). If there is external documentation on the
cryptographic engines in use please include those references (e.g. OpenSSL)
 -->

<!-- write a short text on the technologies used -->

### Encryption of data in transit
<!-- SSDLC:
Canonical Kubernetes incorporates a variety of cryptographic algorithms to
secure internal operations and data at multiple levels. The primary technologies
are as follows:
TLS (Transport Layer Security)
Protocols Supported: TLS 1.2 and TLS 1.3 are employed for encryption of data in
transit. TLS is the standard for securing connections and is used across all
Canonical Kubernetes components.
Symmetric Encryption Algorithm: AES-GCM, a widely adopted and secure symmetric
encryption method, is utilized to ensure confidentiality and integrity of data.
Key Length: AES-256, providing a high level of security suitable for sensitive
communications.
Asymmetric Algorithms:
RSA-2048 and RSA-4096: Canonical Kubernetes utilizes RSA for key exchanges and
digital signatures, providing compatibility with a broad range of cryptographic
libraries and external systems.
ECDSA: In environments requiring higher performance with lower computational
overhead, Canonical Kubernetes supports ECDSA, particularly beneficial for
devices with constrained resources.
Hashing Algorithms:
SHA-256: Used for hashing and ensuring data integrity, SHA-256 is a
cryptographic hash function providing a strong safeguard against data tampering.
 -->

### Encryption of data at rest

<!-- SSDLC:
Dqlite Encryption at Rest
Encryption Algorithm: AES-GCM is used for data encryption within Dqlite, which
manages database storage securely. This encryption ensures that sensitive
information, including Kubernetes Secrets, remains protected from unauthorized
access at rest.
Key Management: Key lengths for AES encryption in Dqlite typically utilize
AES-256 for robust data security. -->

## Configure cryptography in {{product}}
<!-- Guidance:
Cryptographic technology being exposed to the user for their use
In this section list/describe the cryptographic technology that is exposed for
users to utilize. Be sure to include algorithms and key lengths supported. If
there is external documentation on the cryptographic engines we’re exposing
please include those references (e.g. OpenSSL) -->

<!-- SSDLC:
Canonical Kubernetes provides various cryptographic tools that users can
leverage to implement security controls for their workloads:
Kubernetes Secrets Encryption: Users are empowered to configure encryption
providers for Secrets at rest, with AES-GCM as the preferred encryption
algorithm for data confidentiality.
Configurable Providers: Canonical Kubernetes supports integration with various
Key Management Services (KMS) to centralize and control key management.
Kubernetes API Authentication: For secure API access, users can configure X.509
certificates, allowing them to implement secure, certificate-based
authentication with support for both RSA (2048/4096) and ECDSA keys.
Service Mesh Encryption (Optional): When deploying service meshes like Istio
with Canonical Kubernetes, users can enable mutual TLS (mTLS) to protect
inter-service communications, ensuring data privacy and authenticity in
multi-service environments.
Supported Algorithms for mTLS:
RSA-2048/4096 or ECDSA: These options enable secure certificate-based
authentication between services.
AES-GCM: Used for encrypted service-to-service communication.
 -->

## Third party cryptographic packages and libraries
<!-- Guidance:
Packages or technology providing cryptographic functionality. For example,
if the functionality is being supplied by the linux kernel, or openssl from the
Ubuntu Archive, please list this. If any packages or software from outside of
Ubuntu are being used, please also provide the repository they are being sourced
from. -->

<!-- SSDLC:
Canonical Kubernetes depends on a suite of cryptographic libraries and
packages to implement its security functions:
OpenSSL: Canonical Kubernetes utilizes OpenSSL for a broad range of c
ryptographic operations, including TLS, certificate management, and secure key
exchange. OpenSSL’s extensive cryptographic functionality and secure algorithms
provide a reliable foundation for TLS and encryption operations.
Source: Ubuntu Archive (Package: openssl)
Linux Kernel Cryptographic Modules: For network security and cryptographic
operations at the kernel level, Canonical Kubernetes leverages cryptographic
modules in the Linux kernel, such as IPsec for secure network communications.
Source: Ubuntu Kernel (Package: linux-generic)
Go Cryptography Library: Since Kubernetes is written in Go, Canonical Kubernetes
relies on the Go standard library's cryptographic functions, implementing secure
algorithms such as RSA, ECDSA, and AES, which are necessary for the secure
operation of Kubernetes components.
Source: Go Standard Library -->

## Recommended usage and settings
<!-- leave as default / set up certain components / link to certificates how-tos
could link to usage section above -->