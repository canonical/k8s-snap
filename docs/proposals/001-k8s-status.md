
# Proposal information

<!-- Index number -->
- **Index**: 001

<!-- Status -->
- **Status**: **DRAFTING**

<!-- Short description for the feature -->
- **Name**: Expose feature status in `k8s status` command

<!-- Owner name and github handle -->
- **Owner**: Angelos Kolaitis [@neoaggelos](https://github.com/neoaggelos)

# Proposal Details

## Summary
<!--
In a short paragraph, explain what the proposal is about and what problem
it is attempting to solve.
-->

The `k8s status` command should expose the _status_ of the built-in Canonical
Kubernetes features.

This can be used to allow users to verify that all built-in features of
Canonical Kubernetes (e.g. `network`, `dns`, etc) are deployed on the cluster.

Otherwise, if an error has occured (e.g. configuration is not valid, or some
resource could not be deployed), this error will be raised to the user as well.

## Rationale
<!--
This section COULD be as short or as long as needed. In the appropriate amount
of detail, you SHOULD explain how this proposal improves k8s-snap, what is the
problem it is trying to solve and how this makes the user experience better.

You can do this by describing user scenarios, and how this feature helps them.
You can also provide examples of how this feature may be used.
-->

The current `k8s status` command returns information about the control plane
nodes, and then mimics the output of the `k8s get` command. This is not useful,
as this is by definition the "spec" of what the built-in features should be, not
their status.

If anything happens and deploying a built-in feature fails, then `k8s status`
does not present this information in any way to the user. Further, the only way
to debug this is to look for logs in the `k8sd` service, which users might not
be familiar with.

Instead, we want the `k8s status` command to return meaningful information about
the current _deployment status_ of each feature. We want to replace the
following output (taking the `dns` feature as an example):

```
$ k8s status
...
dns:
  enabled: true                         # these fields comes from the cluster
  service-ip: 10.152.183.10             # configuration. enabled=true does not
  cluster-domain: cluster.local         # mean that 'dns' is actually available
```

With something like this:

```
$ k8s status
...
dns:
  enabled: true/false                   # true if successful, false otherwise
  message: CoreDNS deployed
```

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->

### `k8s status` command

The output format of `k8s status` will change.

#### Old output

```
$ k8s status
status: ready
high-availability: no
datastore:
  type: k8s-dqlite
  voter-nodes:
    - 10.87.24.119:6400
  standby-nodes: none
  spare-nodes: none
network:
  enabled: true
dns:
  enabled: true
  cluster-domain: cluster.local
  service-ip: 10.152.183.104
  upstream-nameservers:
  - /etc/resolv.conf
ingress:
  enabled: false
  default-tls-secret: ""
  enable-proxy-protocol: false
load-balancer:
  enabled: false
  cidrs: []
  l2-mode: false
  l2-interfaces: []
  bgp-mode: false
  bgp-local-asn: 0
  bgp-peer-address: ""
  bgp-peer-asn: 0
  bgp-peer-port: 0
local-storage:
  enabled: true
  local-path: /var/snap/k8s/common/rawfile-storage
  reclaim-policy: Delete
  default: true
gateway:
  enabled: true
```

#### New output

With `--output-format=plain`, a simple overview is shown, with `yaml` or `json`
more fields are shown:

```
$ k8s status
status: ready
high-availability: no
datastore:
  type: k8s-dqlite
  voter-nodes:
    - 10.87.24.119:6400
  standby-nodes: none
  spare-nodes: none
network:
  enabled: true
dns:
  enabled: true
ingress:
  enabled: false
load-balancer:
  enabled: false
local-storage:
  enabled: true
gateway:
  enabled: false
  message: Cilium GatewayAPI support could not be enabled, error was "..."
```

When the output format is `json` or `yaml`, more fields may be shown:

```
$ k8s status --output-format=yaml
...
network:
  enabled: true
  message: Cilium is deployed
  timestamp: 2020-01-01 10:00:00
  version: v1.15.2
...
```

## Alternative solutions
<!--
This section SHOULD list any possible alternative solutions that have been or
should be considered. If required, add more details about why these alternative
solutions were discarded.
-->

none

## Out of scope
<!--
This section MUST reference any work that is out of scope for this proposal.
Out of scope items are typically unknowns that we do not yet have a clear idea
of how to solve, so we explicitly do not tackle them until we have more
information.

This section is very useful to help guide the implementation details section
below, or serve as reference for future proposals.
-->

### `message` field

The `message` field is the only information that will be presented to the user.
We should be explicit that this field is only meant to be informative, users
should not try to programmatically parse it to extract meaningful information
about the feature.

In case things are working as expected, the message could be empty (such that
the output of `k8s status` remains concise).

It is out-of-scope for this proposal to provide any sort of _structured_
fields into the output of `k8s status` command.

### `version` field

This field is set to the application version of the feature (e.g. `v1.15.2`).
In the future, this may be used to detect when we _can_ or _should_ upgrade
the version of features (e.g. a newer Canonical Kubernetes version would come
with Cilium v1.16.2), we could add a similar `upgradeable-to` field.

This is out of scope for this proposal.

### `enabled` field

For the output of `k8s status`, we only care that the feature itself has been
successfully deployed on the cluster. When this field is true, it means that any
manifests required were successfully deployed on the cluster.

In the future, we might want to extend this with a `ready: true/false` field,
indicating that the feature is ready and/or available and/or healthy.

This is out of scope for this proposal.

# Implementation Details

## API Changes
<!--
This section MUST mention any changes to the k8sd API, or any additional API
endpoints (and messages) that are required for this proposal.

Unless there is a particularly strong reason, it is preferable to add new v2/v3
APIs endpoints instead of breaking the existing APIs, such that API clients are
not affected.
-->

Add a new type in [api/v1/types.go]

```go
type FeatureStatus struct {
    Enabled bool
    Message string
    Version string
    Timestamp time.Time
}
```

Extend [apiv1.ClusterStatus] with fields for the status of individual features.

```go
type ClusterStatus struct {
    // ...
    Network FeatureStatus
    DNS FeatureStatus
}
```

Extend [pkg/k8sd/types] with the internal `FeatureStatus` struct (this should be
a separate struct, such that internal types and API can evolve separately).

```go
type FeatureStatus struct {
    Enabled bool
    Message string
    Version string
    Timestamp time.Time
}
```

Two helper methods should also be added to convert between internal<->API types:

```go
func (f FeatureStatus) ToAPI() apiv1.FeatureStatus
func FeatureStatusFromAPI(apiv1.FeatureStatus) FeatureStatus
```

## CLI Changes
<!--
This section MUST mention any changes to the k8s CLI, e.g. new arguments,
different outputs.
-->

These changes will affect the output of the `k8s status` command, as explained
in the section above.

## Database Changes
<!--
This section MUST mention any changes required in the k8sd database schema or
internal types.
-->

We need to define a new table (changes go in [pkg/k8sd/database]):

```sql
CREATE TABLE feature_status (
    id          INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    name        TEXT UNIQUE NOT NULL,
    message     TEXT NOT NULL,
    version     TEXT NOT NULL,
    timestamp   TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL,
    UNIQUE(name)
)
```

We then need to add the following methods to our model:

```go
// SetFeatureStatus updates the status of an individual feature.
SetFeatureStatus(ctx context.Context, tx *sql.Tx, name string, status types.FeatureStatus) error

// GetFeatureStatuses returns a map(featureName => types.FeatureStatus).
GetFeatureStatuses(ctx context.Context, tx *sql.Tx) (map[string]types.FeatureStatus, error)
```

## Configuration Changes
<!--
This section MUST mention any new configuration options or service arguments
that are introduced.
-->
none

## Documentation Changes
<!--
This section MUST mention any new documentation that is required for the new
feature. Most features are expected to come with at least a How-To and an
Explanation page.

In this section, it is useful to think about any existing pages that need to be
updated (e.g. command outputs).
-->

We must update all references and example of `k8s status` outputs to reflect
the new output format.

## Testing
<!--
This section MUST explain how the new feature will be tested.
-->

- Unit tests for changes in [pkg/k8sd/database].
- Unit tests for changes in [pkg/k8sd/features] (see below).
- Manual testing of `k8s status`, such that output matches our expectations.
- Extend [tests/integration/tests/test_smoke.py] with a basic test that the
  output of `k8s status` matches expectations.

## Considerations for backwards compatibility
<!--
In this section, you MUST mention any breaking changes that are introduced by
this feature. Some examples:

- In case of deleting a database table, how do older k8sd instances handle it?
- In case of a changed API endpoint, how do existing clients handle it?
- etc
-->

We extend the [apiv1.ClusterStatus] type with new fields. These will be empty
when decoding responses from older k8sd versions. This should not be cause for
trouble, as `k8s` and `k8sd` versions always match.

API consumers that depend on the previous response will still work, as we are
not taking away the Config field.

## Implementation notes and guidelines
<!--
In this section, you SHOULD go into detail about how the proposal can be
implemented. If needed, link to specific parts of the code (link against
particular commits, not branches, such that any links remain valid going
forward).

This is useful as it allows the proposal owner to not be the person that
implements it.
-->

We'll use the example of `features.ApplyNetwork()` as an example. Similar things
are required for the rest of the features:

Update the function signature in [pkg/k8sd/features.Interface]:

```go
// Return a types.FeatureStatus object along with the error.
// The `types.FeatureStatus` must be a valid object in all cases (even if we
// return an error).
// In case of error, it should contain a wrapped form of the error message:
// "Failed to deploy Cilium, the error was %w"
// Callers test for success or failure by checking for err == nil.
func ApplyNetwork(context.Context, snap.Snap, types.Network, types.Annotations) (types.FeatureStatus, error)
```

Features are managed by [pkg/k8sd/controllers.FeatureController], and in
particular the [(*FeatureController).reconcileLoop]. There, two changes are
needed:

1. Pass an extra parameter to [(*FeatureController).Run], a callback function
   that updates the feature status, which should look something like:

   ```go
   setFeatureStatus func(context.Context, name string, status types.FeatureStatus) error
   ```

   That would be implemented when starting the controller in [(*App).onStart],
   by capturing the `*state.State` and calling the database model methods.

2. Adjust [(*FeatureController).reconcileLoop] to call `setFeatureStatus` after
   running the `apply` method.

Finally, we need to extend [(*Endpoints).getClusterStatus] with a call to the
`database.GetFeatureStatuses()` and setting the individual fields on the
response, e.g.

```go
statuses, err := database.GetFeatureStatuses(ctx, ...)
response.NetworkStatus = statuses["network"]
response.DNSStatus = statuses["dns"]
// ....
```

<!-- LINKS -->
[api/v1/types.go]: https://github.com/canonical/k8s-snap/blob/9c260479d7201f231817ee95131444c534f29c33/src/k8s/api/v1/types.go
[apiv1.ClusterStatus]: https://github.com/canonical/k8s-snap/blob/9c260479d7201f231817ee95131444c534f29c33/src/k8s/api/v1/types.go#L51
[pkg/k8sd/database]: https://github.com/canonical/k8s-snap/tree/9c260479d7201f231817ee95131444c534f29c33/src/k8s/pkg/k8sd/database
[pkg/k8sd/types]: https://github.com/canonical/k8s-snap/tree/9c260479d7201f231817ee95131444c534f29c33/src/k8s/pkg/k8sd/types
[pkg/k8sd/features]: https://github.com/canonical/k8s-snap/tree/9c260479d7201f231817ee95131444c534f29c33/src/k8s/pkg/k8sd/types
[tests/integration/tests/test_smoke.py]: https://github.com/canonical/k8s-snap/blob/9c260479d7201f231817ee95131444c534f29c33/tests/integration/tests/test_smoke.py
[pkg/k8sd/features.Interface]: https://github.com/canonical/k8s-snap/blob/f026b18c18a4e80d5dd0b6147ef6a638127d415d/src/k8s/pkg/k8sd/features/interface.go#L15
[(*FeatureController).reconcileLoop]: https://github.com/canonical/k8s-snap/blob/f026b18c18a4e80d5dd0b6147ef6a638127d415d/src/k8s/pkg/k8sd/controllers/feature.go#L123
[(*App).onStart]: https://github.com/canonical/k8s-snap/blob/4c5f44934e1927976f28d532519452081bee9321/src/k8s/pkg/k8sd/app/hooks_start.go#L55-L74
[(*Endpoints).getClusterStatus]: https://github.com/canonical/k8s-snap/blob/7cd6466232c9b001edeaefb472d04e5da34237fd/src/k8s/pkg/k8sd/api/cluster.go#L14
