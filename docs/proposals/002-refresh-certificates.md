<!--
To start a new proposal, create a copy of this template on this directory and
fill out the sections below.
-->

# Proposal information

<!-- Index number -->
- **Index**: 002

<!-- Status -->
- **Status**: **DRAFTING** <!-- **DRAFTING**/**ACCEPTED**/**REJECTED** -->

<!-- Short description for the feature -->
- **Name**: Refresh certificates

<!-- Owner name and github handle -->
- **Owner**: Angelos Kolaitis [@neoaggelos](https://github.com/neoaggelos)

# Proposal Details

## Summary
<!--
In a short paragraph, explain what the proposal is about and what problem
it is attempting to solve.
-->

This proposal defines how cluster nodes (control plane or workers) can refresh
their certificates to ensure continuous operation of the cluster.

## Rationale
<!--
This section COULD be as short or as long as needed. In the appropriate amount
of detail, you SHOULD explain how this proposal improves k8s-snap, what is the
problem it is trying to solve and how this makes the user experience better.

You can do this by describing user scenarios, and how this feature helps them.
You can also provide examples of how this feature may be used.
-->

Kubernetes clusters require quite a few TLS certificates for their normal
operation. These certificates are used to serve TLS, or client authentication.

Server certificates are typically signed by the `kubernetes-ca` CA certificate
and client certificates are signed by `kubernetes-client-ca`. CA certificates
are typically long-lived (e.g. 10 years) and should not change except for
extraordinary situations. On the contrary, individual node certificates are
expected to be short-lived and rotated often.

Given that certificate rotation is a fundamental operation for a Kubernetes
cluster, we should have a command akin to `k8s refresh-certs`.

In a default Canonical Kubernetes setup, the control plane nodes of the cluster
own the CA certificate and key, therefore they can sign new certificates for the
nodes to ensure continuous operation of the cluster.

For worker nodes, this is not possible. Instead of going with a custom
communication layer between workers and control plane nodes, we will use the
standard Kubernetes mechanism of [Certificate Signing Requests]. When a worker
node wants to refresh their certificates, they use their existing kubelet config
to create the necessary CSR objects, wait for them to be approved and signed,
and then take the new certificates and refresh their local configs. The are
default [Kubernetes signers] that we could use, but they are not sufficient for
our needs, as we need better control over our certificates (e.g. we also need
kube-proxy client certificates). However, we will re-use this approach to write
a controller running as part of k8sd which handles this for us.

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->
none, existing CLI commands and API endpoints are not affected.

## Alternative solutions
<!--
This section SHOULD list any possible alternative solutions that have been or
should be considered. If required, add more details about why these alternative
solutions were discarded.
-->

### Vault operator

A future direction would be to directly integrate k8sd with an external service
like Vault. The k8sd service on each node could be given a vault URL, token,
identity and the required configs. Then, those credentials may be used by the
node to retrieve its certificates and automatically manage service restarts as
needed.

This would be a great solution for certificate refreshes, but comes with two
main caveats:

1. It requires a lot of work to introduce the concept of `PKIProvider` to k8sd,
   where certificates are not signed using a key but retrieved from an external
   service.
2. It requires an external Vault service to be set up and operated outside of
   k8sd, which is not always desirable or possible.

Therefore, this could be considered as a future direction for environments where
such a setup is possible.

## Out of scope
<!--
This section MUST reference any work that is out of scope for this proposal.
Out of scope items are typically unknowns that we do not yet have a clear idea
of how to solve, so we explicitly do not tackle them until we have more
information.

This section is very useful to help guide the implementation details section
below, or serve as reference for future proposals.
-->

### Refresh certificates across the cluster

The `k8s refresh-certs` command only refreshes the certificates on the current
node. Refreshing the certificates of all cluster nodes poses risks and possibly
unpredictable control plane or workload downtime, therefore is out of scope for
the current proposal.

### Microcluster certificates

microcluster does not provide any mechanism to rotate or refresh the cluster and
server certificates of the nodes. Certificates for each node are created with a
10 year TTL, but there might be environments where this is not allowed.

### Refresh CA certificate

Refreshing the individual node certificates can often be performed "online", as
it typically only requires a restart of the Kubernetes services on the node
(such that the new certificates are picked up). It often comes without much
noticeable user-facing control plane disruption, and does not have any effect on
the cluster workloads.

On the contrary, refreshing the CA certificates of a cluster comes with serious
downtime:

- Client authentication is broken before all nodes are aware of the new CA.
- The Kubernetes CA certificate is visible and used by workloads that must
  interact with the Kubernetes API (e.g. CNI, controllers, etc). These pods
  must then be restarted after the certificates have been refreshed.

Due to the very disruptive nature of the operation, refreshing the cluster CA
is kept intentionally out of scope for the existing proposal, and will be dealt
with separately in the future.

### External CA certificate

When deploying Canonical Kubernetes with an external CA certificate, the cluster
does not know the CA key. Therefore, it is unable to sign new node certificates.

Due to this, it is the responsibility of the orchestration tool that is used to
sign the new certificates and make them available on the cluster nodes.

Automating any part of this manual process is out of scope for this proposal.

# Implementation Details

## API Changes
<!--
This section MUST mention any changes to the k8sd API, or any additional API
endpoints (and messages) that are required for this proposal.

Unless there is a particularly strong reason, it is preferable to add new v2/v3
APIs endpoints instead of breaking the existing APIs, such that API clients are
not affected.
-->

Certificate refreshes behave differently for control plane and worker nodes.
Control plane nodes have access to the CA certificate and private key, therefore
can easily refresh their own certificates. Worker nodes do not have access to
the CA private key, so certificate refreshes need to happen in two steps.

To accomodate this need, while also ensuring that the `k8s refresh-certs` can be
straightforward to implement, two API calls are introduced, which clients should
call in order.

### `POST /refresh-certs/init`

```go
type RefreshCertificatesInitRequest struct {
}

type RefreshCertificatesInitResponse struct {
  // Seed needs to be passed to `/refresh-certs/run` to ensure that CSR
  // names are valid.
  Seed int
  // RequiredCertificateApprovals is a list of CSRs (kubectl get csr) that need
  // to be manually approved when `/refresh-certs/run` is used. This list will
  // be empty for control plane nodes.
  RequiredCertificateApprovals []string
}
```

`POST /refresh-certs/init` performs any necessary preparations for certificate
refreshes on the current node.

It returns a `seed` that must be passed to `POST /refresh-certs/run`. For worker
nodes, it also returns a list of names of CertificateSigningRequest objects that
will need to be approved and signed. This can be used by the CLI to print a
helpful message and request that the certificates be signed.

This method call is purely informative and does not have any side-effect on the
cluster.

### `POST /refresh-certs/run`

```go
type RefreshCertificatesRunRequest struct {
  // Seed must match the seed from the received RefreshCertificatesInitResponse.
  Seed int
  // ExpirationDuration is the duration of the requested certificates.
  ExpirationDuration time.Duration
}

type RefreshCertificatesRunResponse struct {
  // ExpirationDuration is the duration of the new certificates.
  ExpirationDuration time.Duration
}
```

#### Worker nodes

Worker nodes will create the required CSRs, and wait for them to be signed. This
comprises of two steps: first they must be approved (see below), and then the
csrsigning controller will sign them. The worker node will block until the CSRs
are dealt with:

- If any CSR becomes Failed or Denied, then we fail and return the result.
- If any CSR does not become Approved or Issued in time, we time out.

After getting the new certificates, we apply them on the node, do any service
restarts that may be required and return response.

#### Control plane nodes

Control plane nodes can ignore the seed argument. They already have access to
the CA certificates and keys, therefore can simply generate and sign new
certificates, apply them on the node, do any service restarts necessary and
return the response.

## CLI Changes
<!--
This section MUST mention any changes to the k8s CLI, e.g. new arguments,
different outputs.
-->

### `k8s refresh-certs` command

#### For worker nodes

```
root@worker:~$ k8s refresh-certs --ttl 30d --timeout 30s
The following CertificateSigningRequests should be approved. Run the following
commands on any of the control plane nodes of the cluster:

  $ k8s kubectl certificate approve k8sd-seed-worker-kubelet-client
  $ k8s kubectl certificate approve k8sd-seed-worker-kubelet-serving
  $ k8s kubectl certificate approve k8sd-seed-worker-kube-proxy-client

Waiting for certificates to be created....

<-- CSRs are approved -->

Certificates have been refreshed, and will expire at $date.
```

#### For control plane nodes

```
root@node:~$ k8s refresh-certs --ttl 30d --timeout 30s
Certificates have been refreshed, and will expire at $date.
```

## Database Changes
<!--
This section MUST mention any changes required in the k8sd database schema or
internal types.
-->
none

## Configuration Changes
<!--
This section MUST mention any new configuration options or service arguments
that are introduced.
-->

The changes include a `csrsigning` controller, which watches the cluster for
`CertificateSigningRequest` objects coming from the worker nodes. After they are
approved, it will sign the certificates.

### k8sd `--disable-csrsigning-controller`

In line with the rest of the `--disable-$foo-controller` flags, this flag can be
used to completely disable the csrsigning controller on a node.

### annotation `k8sd/v1alpha1/csrsigning/auto-approve=true`

In order to make operations simpler for cases where administrators are not too
strict about security, we can add an annotation that allows the `csrsigning`
controller to automatically approve CertificateSigningRequests that originate
from worker nodes.

This can be configured by setting the annotation as bootstrap config, or
dynamically with:

```bash
k8s set annotations=k8sd/v1alpha1/csrsigning/auto-approve=true      # enable
k8s set annotations=k8sd/v1alpha1/csrsigning/auto-approve=false     # disable
```

## Documentation Changes
<!--
This section MUST mention any new documentation that is required for the new
feature. Most features are expected to come with at least a How-To and an
Explanation page.

In this section, it is useful to think about any existing pages that need to be
updated (e.g. command outputs).
-->

1. We need an explanation page that talks about certificate refreshes. What it
   is, what it is used for and how it helps with cluster security.
2. We need a how-to that describes how to refresh certificates on a cluster with
   N control plane nodes and M worker nodes.
3. We need a reference page with the used cluster certificates, their subject,
   their usage, their path on the disk, as well as what they are used for.

## Testing
<!--
This section MUST explain how the new feature will be tested.
-->

## Considerations for backwards compatibility
<!--
In this section, you MUST mention any breaking changes that are introduced by
this feature. Some examples:

- In case of deleting a database table, how do older k8sd instances handle it?
- In case of a changed API endpoint, how do existing clients handle it?
- etc
-->
none

## Implementation notes and guidelines
<!--
In this section, you SHOULD go into detail about how the proposal can be
implemented. If needed, link to specific parts of the code (link against
particular commits, not branches, such that any links remain valid going
forward).

This is useful as it allows the proposal owner to not be the person that
implements it.
-->


<!-- LINKS -->
[Certificate Signing Requests]: https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/
[Kubernetes signers]: https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/#kubernetes-signers
