# FIPS and cryptography in k8s-snap

This document provides a comprehensive overview of FIPS 140-3 compliance for
the Canonical Kubernetes snap, covering both snap binaries and ROCK images. It
details which components contain cryptographic functionality, how they are built
to support FIPS compliance, and the mechanisms that enable FIPS-validated
cryptography at runtime.

## Go toolchain

The k8s-snap achieves FIPS compliance by using a modified Go toolchain from the
[go snap](https://snapcraft.io/go). This snap is based on [Microsoft's Go
fork](https://github.com/microsoft/go/) with [snap-specific
modifications](https://bugs.launchpad.net/go-snap).

The modified Go toolchain fundamentally changes how Go's crypto packages
operate: instead of using Go's native cryptographic implementations, the runtime
is modified to delegate all cryptographic operations to the host system's
OpenSSL library. This is enabled through the `GOEXPERIMENT=opensslcrypto` build
flag. The toolchain dynamically links against OpenSSL at runtime, which means:

1. **Host Dependency**: The binaries require a FIPS-enabled host system with
   FIPS-validated OpenSSL modules
2. **Automatic FIPS Mode**: When `/proc/sys/crypto/fips_enabled=1` on the host,
   OpenSSL automatically uses its FIPS-validated crypto modules
3. **No Code Changes**: Applications built with this toolchain transparently use
   FIPS-compliant crypto without source code modifications
4. **Dynamic Linking Required**: Binaries containing cryptographic code must be
   built with `CGO_ENABLED=1` and dynamically linked to enable runtime
   delegation to OpenSSL

For complete technical details on how FIPS works with this modified Go
toolchain, see the [Microsoft Go FIPS
documentation](https://github.com/microsoft/go/blob/microsoft/main/eng/doc/fips/README.md).

The k8s-snap leverages FIPS-validated OpenSSL modules from the Ubuntu base snap,
which provides OpenSSL 3.x compiled with [FIPS support on FIPS
channels](https://ubuntu.com/tutorials/using-the-ubuntu-pro-client-to-enable-fips#1-overview).

## Binaries

| Binary                                    | Build Type | Crypto | Notes                                                                                                                                             |
|-------------------------------------------|------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| [cni][cni-mod]                            | Dynamic    | ✅      | TLS for network policy enforcement, certificate handling, secure communication with network controllers, crypto not in go.mod but vendored in SDK |
| [containerd][containerd-mod]              | Dynamic    | ✅      | TLS for gRPC API communication, image signature verification, registry authentication                                                             |
| [containerd-shim-runc-v2][containerd-mod] | Dynamic    | ✅      | Low-level shim between containerd and runc with crypto for secure communication                                                                   |
| [ctr][containerd-mod]                     | Dynamic    | ✅      | Command-line client with TLS support for secure registry and API communication                                                                    |
| [etcd][etcd-mod]                          | Dynamic    | ✅      | TLS for client-server and peer-to-peer communication, data encryption at rest, client certificate authentication                                  |
| [helm][helm-mod]                          | Dynamic    | ✅      | TLS for chart repository communication, chart signature verification, secure connections to Kubernetes API                                        |
| [k8s-apiserver-proxy][k8sd-mod]           | Dynamic    | ✅      | TLS proxy for API server                                                                                                                         |
| [k8s-dqlite][k8s-dqlite-mod]              | Dynamic    | ✅      | TLS for Dqlite cluster communication, certificate-based authentication between nodes, encrypted replication streams                               |
| [k8sd][k8sd-mod]                          | Dynamic    | ✅      | Cluster management API with TLS                                                                                                                  |
| [kubernetes][kubernetes-mod]              | Dynamic    | ✅      | TLS communications, certificate management, token signing, encryption at rest, authentication (main binary for all kube-* binaries)               |
| [runc][runc-mod]                          | Static     | ❌      | Low-level container runtime with no network communication or cryptographic operations                                                             |

## Adding or updating a component to the k8s-snap

1. **As a developer, check if the component uses any crypto**:

   ```bash
   # Output will be empty if no crypto used
   go list -deps ./... | grep golang.org/x/crypto
   ```

2. **Determine build type**:
   - If crypto is used → Build with `CGO_ENABLED=1`,
     `GOEXPERIMENT=opensslcrypto` (see [FIPS
     docs](https://github.com/microsoft/go/blob/microsoft/main/eng/doc/fips/README.md)
     for details)
   - If no crypto → Can be built statically

3. **Update this document** with the new component's details

4. **Test FIPS compliance** on a FIPS-enabled system (partly done in weekly CI
   runs)

## ROCKs

k8s-snap uses several ROCK images for optional cluster features. These ROCKs
are built with FIPS support using Ubuntu Pro FIPS-validated OpenSSL libraries.
All ROCKs are built with FIPS compliance using:

1. **OpenSSL**: FIPS-validated OpenSSL library included, see [this discourse
   post](https://discourse.ubuntu.com/t/build-rocks-with-ubuntu-pro-services/57578):

   ```yaml
   parts:
     openssl:
       plugin: nil
       stage-packages:
         - openssl-fips-module-3
         - openssl
   ```

2. **Go Toolchain**: Microsoft's modified Go toolchain with
   `GOEXPERIMENT=opensslcrypto`
3. **Build Command**: `rockcraft pack --pro=fips-updates`

During runtime, ROCKs automatically detect FIPS mode when
`/proc/sys/crypto/fips_enabled=1` on the host system and use the bundled
FIPS-validated OpenSSL modules accordingly.

| ROCK                                    | Repository                                                          | FIPS Support Since | Components                          | Notes                                                                                                                                                              |
|-----------------------------------------|---------------------------------------------------------------------|--------------------|-------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [cilium][cilium-rock]                   | [cilium][cilium-repo], [cilium-operator][cilium-repo]               | `1.17.9-ck2`       | cilium-agent, cilium-operator       | Uses TLS for cluster mesh, API server communication, and secure service mesh features                                                                              |
| [coredns][coredns-rock]                 | [coredns][coredns-repo]                                             | `1.12.4-ck1`       | coredns                             | Uses TLS for DNS-over-TLS (DoT), DNS-over-HTTPS (DoH), and secure communication with backend services                                                              |
| [metallb][metallb-rock]                 | [metallb-controller][metallb-repo], [metallb-speaker][metallb-repo] | `v0.14.9-ck4`      | metallb-controller, metallb-speaker | Controller uses TLS for Kubernetes API. Speaker handles BGP/NDP with optional authentication. Note: MD5 authentication is not FIPS-compliant and should be avoided |
| [metrics-server][metrics-server-rock]   | [metrics-server][metrics-server-repo]                               | `0.8.0-ck4`        | metrics-server                      | Uses TLS for secure communication with Kubernetes API and webhook servers, certificate validation                                                                  |
| [rawfile-localpv][rawfile-localpv-rock] | [rawfile-localpv][rawfile-localpv-repo]                             | `0.8.2-ck3`        | rawfile-localpv CSI driver          | Uses TLS for CSI gRPC communication and secure volume provisioning operations                                                                                      |

### ROCK Cryptographic Usage

See each ROCK’s `docs/fips.md` file for a detailed analysis of their
crypto usage.

## Related Documentation

- [User-facing FIPS Installation Guide](../canonicalk8s/snap/howto/install/fips.md)
- [Snapcraft Configuration](../../snap/snapcraft.yaml)
- [Component Build Scripts](../../build-scripts/components/)

## References

- [FIPS 140-3 Standard](https://csrc.nist.gov/pubs/fips/140-3/final)
- [Ubuntu FIPS Documentation](https://ubuntu.com/security/certifications/docs/fips)

<!-- Links -->

[cni-mod]: https://github.com/containernetworking/plugins/blob/main/go.mod
[containerd-mod]: https://github.com/containerd/containerd/blob/main/go.mod
[etcd-mod]: https://github.com/etcd-io/etcd/blob/main/go.mod
[helm-mod]: https://github.com/helm/helm/blob/main/go.mod
[k8sd-mod]: ../../src/k8s/go.mod
[k8s-dqlite-mod]: https://github.com/canonical/k8s-dqlite/blob/master/go.mod
[kubernetes-mod]: https://github.com/kubernetes/kubernetes/blob/master/go.mod
[runc-mod]: https://github.com/opencontainers/runc/blob/main/go.mod

[cilium-rock]: https://github.com/canonical/cilium-rocks
[cilium-repo]: https://github.com/cilium/cilium
[coredns-rock]: https://github.com/canonical/coredns-rock
[coredns-repo]: https://github.com/coredns/coredns
[metallb-rock]: https://github.com/canonical/metallb-rocks
[metallb-repo]: https://github.com/metallb/metallb
[metrics-server-rock]: https://github.com/canonical/metrics-server-rock
[metrics-server-repo]: https://github.com/kubernetes-sigs/metrics-server
[rawfile-localpv-rock]: https://github.com/canonical/rawfile-localpv
[rawfile-localpv-repo]: https://github.com/openebs/rawfile-localpv
