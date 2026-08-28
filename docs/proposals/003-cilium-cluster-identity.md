
# Proposal information

<!-- Index number -->
- **Index**: 003

<!-- Status -->
- **Status**: **DRAFTING**

<!-- Short description for the feature -->
- **Name**: Cilium ClusterMesh identity via annotations

<!-- Owner name and github handle -->
- **Owner**: <!-- Owner name / [@name](https://github.com/name) -->

# Proposal Details

## Summary
<!--
In a short paragraph, explain what the proposal is about and what problem
it is attempting to solve.
-->

This proposal introduces two new Cilium annotations,
`k8sd/v1alpha1/cilium/cluster-id` and `k8sd/v1alpha1/cilium/cluster-name`,
which k8sd maps to the Cilium Helm chart values `cluster.id` and
`cluster.name` for the `ck-network` Helm release. These allow users to set
the Cilium ClusterMesh identity of the cluster, a prerequisite for joining
a ClusterMesh.

## Rationale
<!--
This section COULD be as short or as long as needed. In the appropriate amount
of detail, you SHOULD explain how this proposal improves k8s-snap, what is the
problem it is trying to solve and how this makes the user experience better.

You can do this by describing user scenarios, and how this feature helps them.
You can also provide examples of how this feature may be used.
-->

Cilium ClusterMesh requires every member cluster to have a unique cluster
ID (an integer in the range 1-255) and a non-default cluster name. Today,
k8sd renders the `ck-network` Helm release (Cilium) with the upstream
chart defaults for these values (`cluster.id=0`, `cluster.name=default`).
This means that any two Canonical Kubernetes clusters deployed with their
default configuration share the same identity, and attempting to form a
ClusterMesh between them fails or behaves unpredictably.

By exposing `cluster.id` and `cluster.name` through the annotations
mechanism — the established pattern for Cilium tunables in k8sd — users
can assign a proper ClusterMesh identity to each cluster, both at
bootstrap time (via the bootstrap config file) and later via
`k8s set annotations`.

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->

Two new Cilium annotations become available:

- `k8sd/v1alpha1/cilium/cluster-id`
- `k8sd/v1alpha1/cilium/cluster-name`

They can be set like any other annotation, with multiple annotations
comma-separated:

```
sudo k8s set annotations='k8sd/v1alpha1/cilium/cluster-id=1,k8sd/v1alpha1/cilium/cluster-name=my-cluster'
```

At bootstrap time, they can be provided under `cluster-config.annotations`
in the bootstrap config file.

Validation is performed by k8sd, and the following inputs are rejected:
- Invalid cluster ID: not an integer, or outside the `0-255` range.
- Invalid cluster name: empty, longer than 32 characters, contains
  characters outside `[a-z0-9-]`, or has a leading/trailing dash.
- A non-zero cluster ID combined with a default or empty cluster name.

A failed validation results in a reconciliation error, which is surfaced to
the user in the output of `k8s status` (in the `network` feature message).

## Alternative solutions
<!--
This section SHOULD list any possible alternative solutions that have been or
should be considered. If required, add more details about why these alternative
solutions were discarded.
-->

### Typed network config fields

Instead of annotations, `cluster-id` and `cluster-name` could be added as
typed fields of the network config (similar to other network feature
fields).

Rejected: this is a bigger API change, and annotations are the established
pattern in k8sd for Cilium tunables.

### Bootstrap-only fields

The cluster identity could be only configurable at bootstrap time, through
dedicated fields in the bootstrap config file.

Rejected: this reduces flexibility for pre-existing clusters. In any case,
annotations already work at bootstrap time (under
`cluster-config.annotations`).

## Out of scope
<!--
This section MUST reference any work that is out of scope for this proposal.
Out of scope items are typically unknowns that we do not yet have a clear idea
of how to solve, so we explicitly do not tackle them until we have more
information.

This section is very useful to help guide the implementation details section
below, or serve as reference for future proposals.
-->

- ClusterMesh setup itself (connecting clusters, enabling the
  `clustermesh-apiserver` deployment, etc.) is out of scope. This proposal
  only provides the cluster identity values that ClusterMesh requires.
- A generic Helm values passthrough for other Cilium chart values is out of
  scope.

# Implementation Details

## API Changes
<!--
This section MUST mention any changes to the k8sd API, or any additional API
endpoints (and messages) that are required for this proposal.

Unless there is a particularly strong reason, it is preferable to add new v2/v3
APIs endpoints instead of breaking the existing APIs, such that API clients are
not affected.
-->

Two new annotation constants are added to the
`api/annotations/cilium/annotations.go` file in the
[github.com/canonical/k8s-snap-api](https://github.com/canonical/k8s-snap-api)
repository (candidate for the v2.2.0 API version):

```go
// AnnotationClusterID is the annotation for the Cilium cluster ID.
AnnotationClusterID = "k8sd/v1alpha1/cilium/cluster-id"

// AnnotationClusterName is the annotation for the Cilium cluster name.
AnnotationClusterName = "k8sd/v1alpha1/cilium/cluster-name"
```

No k8sd API endpoints are added or modified.

## CLI Changes
<!--
This section MUST mention any changes to the k8s CLI, e.g. new arguments,
different outputs.
-->

None. The annotations are set through the unchanged
`k8s set annotations=...` command.

## Database Changes
<!--
This section MUST mention any changes required in the k8sd database schema or
internal types.
-->

None. Annotations are already stored as part of the cluster configuration.

## Configuration Changes
<!--
This section MUST mention any new configuration options or service arguments
that are introduced.
-->

None.

## Documentation Changes
<!--
This section MUST mention any new documentation that is required for the new
feature. Most features are expected to come with at least a How-To and an
Explanation page.

In this section, it is useful to think about any existing pages that need to be
updated (e.g. command outputs).
-->

The annotations reference page
(`docs/canonicalk8s/snap/reference/annotations.md`) gets two new entries,
one for each new annotation. A ClusterMesh how-to page may be added later,
once ClusterMesh setup itself is supported.

## Testing
<!--
This section MUST explain how the new feature will be tested.
-->

- Unit tests in k8sd covering validation and rendering, in
  `pkg/k8sd/features/cilium/internal_test.go` and
  `pkg/k8sd/features/cilium/network_test.go` (already written alongside the
  implementation).
- Integration test in
  [tests/integration/tests/test_networking.py](https://github.com/canonical/k8s-snap/blob/master/tests/integration/tests/test_networking.py),
  asserting the rendered `ck-network` configmap contains the expected
  `cluster.id` / `cluster.name` values.

## Considerations for backwards compatibility
<!--
In this section, you MUST mention any breaking changes that are introduced by
this feature. Some examples:

- In case of deleting a database table, how do older k8sd instances handle it?
- In case of a changed API endpoint, how do existing clients handle it?
- etc
-->

The defaults are unchanged when the annotations are unset: no `cluster`
values map is rendered into the Helm values unless the annotations are
present, so the chart falls back to its upstream defaults
(`cluster.id=0`, `cluster.name=default`). Existing clusters with no
identity annotations set are unaffected.

## Implementation notes and guidelines
<!--
In this section, you SHOULD go into detail about how the proposal can be
implemented. If needed, link to specific parts of the code (link against
particular commits, not branches, such that any links remain valid going
forward).

This is useful as it allows the proposal owner to not be the person that
implements it.
-->

The change spans three repositories:

- [github.com/canonical/k8s-snap-api](https://github.com/canonical/k8s-snap-api):
  new constants in `api/annotations/cilium/annotations.go` (link to
  specific commit to be linked at merge time).
- [github.com/canonical/k8sd](https://github.com/canonical/k8sd):
  - `pkg/k8sd/features/cilium/internal.go` — validation functions
    `validateClusterID` (accepts `0-255`) and `validateClusterName`.
  - `pkg/k8sd/features/cilium/network.go` — renders the `cluster` values
    map into the Cilium Helm values, guarded by annotation presence.
  - `pkg/k8sd/features/cilium/chart.go` — `ciliumDefaultClusterName`
    constant.
  - (links to specific commits to be linked at merge time)
- [github.com/canonical/k8s-snap](https://github.com/canonical/k8s-snap)
  (this repo): this proposal document and the integration test.

During development, k8sd used a `replace` directive in `go.mod` pointing to
a local k8s-snap-api checkout. This directive must be removed before
merging.
