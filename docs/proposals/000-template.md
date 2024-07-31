<!--
To start a new proposal, create a copy of this template on this directory and
fill out the sections below.
-->

# Proposal information

<!-- Index number -->
- **Index**: 000

<!-- Status -->
- **Status**: <!-- **DRAFTING**/**ACCEPTED**/**REJECTED** -->

<!-- Short description for the feature -->
- **Name**: Feature name

<!-- Owner name and github handle -->
- **Owner**: FirstName LastName / <!-- [@name](https://github.com/name) -->

# Proposal Details

## Summary
<!--
In a short paragraph, explain what the proposal is about and what problem
it is attempting to solve.
-->

## Rationale
<!--
This section COULD be as short or as long as needed. In the appropriate amount
of detail, you SHOULD explain how this proposal improves k8s-snap, what is the
problem it is trying to solve and how this makes the user experience better.

You can do this by describing user scenarios, and how this feature helps them.
You can also provide examples of how this feature may be used.
-->

## User facing changes
<!--
This section MUST describe any user-facing changes that this feature brings, if
any. If an API change is required, the affected endpoints MUST be mentioned. If
the output of any k8s command changes, the difference MUST be mentioned, with a
clear example of "before" and "after".
-->

none

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

none

# Implementation Details

## API Changes
<!--
This section MUST mention any changes to the k8sd API, or any additional API
endpoints (and messages) that are required for this proposal.

Unless there is a particularly strong reason, it is preferable to add new v2/v3
APIs endpoints instead of breaking the existing APIs, such that API clients are
not affected.
-->
none

## CLI Changes
<!--
This section MUST mention any changes to the k8s CLI, e.g. new arguments,
different outputs.
-->
none

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
none

## Documentation Changes
<!--
This section MUST mention any new documentation that is required for the new
feature. Most features are expected to come with at least a How-To and an
Explanation page.

In this section, it is useful to think about any existing pages that need to be
updated (e.g. command outputs).
-->
none

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

## Implementation notes and guidelines
<!--
In this section, you SHOULD go into detail about how the proposal can be
implemented. If needed, link to specific parts of the code (link against
particular commits, not branches, such that any links remain valid going
forward).

This is useful as it allows the proposal owner to not be the person that
implements it.
-->
