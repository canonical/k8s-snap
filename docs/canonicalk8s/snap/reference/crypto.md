# {{product}} cryptography

This reference page provides an overview of the {{product}} cryptography
posture.

### TLS certificates

{{product}} uses TLS secure communications between core components like the api
server and kubelets. By default, {{product}} uses self-signed certificates, but
users are able to use an intermediate CA instead.

{{product}} uses TLS 1.2 and 1.3 for encryption of data in transit. AES-256-GCM
is used to ensure confidentiality and integrity of the data. In addition,
{{product}} uses RSA-2048 and RSA-4096 for key exchanges and digital
signatures. For resource-constrained deployments, {{product}} supports ECDSA.

### Dqlite encryption at rest

{{product}} uses AES-256-GCM (Advanced Encryption Standard - Galois/Counter
Mode) to encrypt cluster data at rest. This is particularly important for
protecting sensitive configurations and credentials stored within the cluster.

### Digital signatures

To authenticate API clients securely, Canonical Kubernetes supports X.509
certificates, leveraging both RSA and ECDSA (Elliptic Curve Digital Signature
Algorithm) for digital signatures, ensuring robust authentication mechanisms.

### Technologies in use

{{product}} uses a variety of open source libraries to implement encryption and
authentication features. OpenSSL is used for certificate management and key
exchange. At the kernel level {{product}} uses Linux kernel cryptographic
modules such as IPsec. Finally, {{product}} relies on the Go standard library's
cryptographic modules for RSA, ECDSA and AES implementations.
