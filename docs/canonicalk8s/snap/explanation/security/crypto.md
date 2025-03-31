# Cryptography

This reference page provides an overview of the {{product}} cryptography
posture.

### TLS certificates

All communications between core components, such as the API server and
kubelets, are encrypted with TLS 1.3 (Transport Layer Security), providing robust protection for sensitive data in transit. By default, {{product}} uses self-signed certificates, but users are able to use an intermediate CA instead.

### Dqlite encryption at rest

{{product}} uses AES-256-GCM (Advanced Encryption Standard - Galois/Counter
Mode) to encrypt cluster data at rest.

### Digital signatures

To securely authenticate API clients, Canonical Kubernetes uses X.509 certificates with support for both RSA-2048 and ECDSA (Elliptic Curve Digital Signature Algorithm). This ensures strong, standards-based authentication with options suited for both general-purpose and resource-constrained environments.

## Configure cryptography in {{product}}

Canonical Kubernetes provides various cryptographic tools that users can
leverage to implement security controls for their workloads:

* Kubernetes Secrets Encryption: Users are empowered to configure encryption
providers for Secrets at rest, with AES-GCM as the preferred encryption
algorithm for data confidentiality.

* Configurable Providers: Canonical Kubernetes supports integration with
various Key Management Services (KMS) to centralize and control key management.

* Kubernetes API Authentication: For secure API access, users can configure X.509 certificates, allowing them to implement secure, certificate-based
authentication with support for RSA 2048 and ECDSA keys.

* Service Mesh Encryption (Optional): When deploying service meshes like Istio
with Canonical Kubernetes, users can enable mutual TLS (mTLS) to protect
inter-service communications, ensuring data privacy and authenticity in
multi-service environments.

  * Supported Algorithms for mTLS:

    * RSA-2048 or ECDSA: These options enable
    secure certificate-based authentication between services.
    * AES-GCM: Used for encrypted service-to-service communication.


## Third party cryptographic packages and libraries

Canonical Kubernetes depends on a suite of cryptographic libraries and packages
to implement its security functions:

* OpenSSL: Canonical Kubernetes utilizes OpenSSL for a broad range of cryptographic operations, including TLS, certificate management, and secure key
exchange. OpenSSL’s extensive cryptographic functionality and secure algorithms
provide a reliable foundation for TLS and encryption operations.

  Source: Ubuntu Archive (Package: [openssl])

* Linux Kernel Cryptographic Modules: For network security and cryptographic
operations at the kernel level, Canonical Kubernetes leverages cryptographic
modules in the Linux kernel, such as IPsec for secure network communications.

  Source: Ubuntu Kernel (Package: [linux-generic])

* Go Cryptography Library: Since Kubernetes is written in Go, Canonical
Kubernetes relies on the Go standard library's cryptographic functions,
implementing secure algorithms such as RSA, ECDSA, and AES, which are necessary
for the secure operation of Kubernetes components.

  Source: Go Standard Library ([x509], [rsa], [sha256], [tls])

## Recommended usage and settings

{{product}} ships with a secure-by-default security posture, so users can rest assured that the default configuration is appropriate for most uses cases. If your security needs are not met by the default configuration, we recommend you [deploy an intermediate CA fine-tuned to your liking][intermediate-ca].

See [certificates] for a list of certificates used in {{product}}.

<!-- LINKS -->

[certificates]: certificates.md
[intermediate-ca]: /snap/howto/security/intermediate-ca.md
[openssl]: https://packages.ubuntu.com/jammy/openssl
[linux-generic]: https://packages.ubuntu.com/jammy/linux-generic
[x509]: https://pkg.go.dev/crypto/x509
[rsa]: https://pkg.go.dev/crypto/rsa
[sha256]: https://pkg.go.dev/crypto/sha256
[tls]: https://pkg.go.dev/crypto/tls
