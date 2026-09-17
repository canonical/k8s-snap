# Proposal information

<!-- Index number -->
- **Index**: 003

<!-- Status -->
- **Status**: **DRAFTING**

<!-- Short description for the feature -->
- **Name**: Automate monthly patch notices

<!-- Owner name and github handle -->
- **Owner**: Niamh Hennigan @nhennigan

# Proposal Details

## Summary

Add a repository-local `patch-notices` documentation tool and scheduled GitHub
Actions workflow to prepare monthly Canonical Kubernetes patch notice updates.
The automation gathers snap and charm release deltas, triages changes with a
configured OpenAI-compatible endpoint, updates release notes and tracking
metadata, and opens a pull request for human review.

## Rationale

Monthly patch notices are currently produced with a mostly manual process:
maintainers identify the latest released revisions for each supported snap and
charm track, inspect the commits since the last documented update, decide which
changes are user-facing, update release notes, and carry state forward for the
next run.

This proposal keeps the human review gate but automates the repetitive data
collection and first-pass summarisation. The workflow creates a reviewable PR
instead of publishing directly, so maintainers can verify generated summaries,
discard false positives, and adjust wording before merge.

## User facing changes

There are no changes to the `k8s` snap, charm, CLI, API, or runtime behaviour.
The only user-facing output is documentation content in the existing release
notes pages under `docs/canonicalk8s/releases/`.

## Alternative solutions

- Continue the fully manual process. This avoids AI and CI integration, but it
  keeps the monthly process slow and easy to run inconsistently across tracks.
- Keep the tool local-only and do not add a scheduled workflow. This helps with
  manual runs, but does not ensure that patch notices are prepared regularly.

## Out of scope

- Publishing or merging patch notice updates without human review.
- Changing the public release notes format beyond adding new dated patch notice
  entries.
- Backport automation changes beyond whatever existing repository backport
  workflows already perform.
- Supporting projects outside Canonical Kubernetes.

# Implementation Details

## API Changes

None. This proposal does not add or change k8sd APIs.

## CLI Changes

No `k8s` CLI changes are introduced.

The repository-local documentation tool adds a `patch-notices` command for
maintainers and CI. The relevant subcommands are:

- `fetch`: collect release delta data for a snap or charm track.
- `review`: run AI triage and write a human-editable workbook.
- `finalize`: parse an approved workbook and update local metadata.
- `generate`: combine fetch, triage, release note insertion, and summary output
  for CI use.
- `pr-body`: build the generated PR body from per-track summaries.

## Database Changes

None. The tool uses a git-tracked JSON metadata file under
`docs/tools/patch-notices/metadata/` to remember the last documented SHA and
date for each track.

## Configuration Changes

The scheduled workflow requires repository secrets for the configured
OpenAI-compatible endpoint and bot token. Secrets should be scoped to the steps
that need them, and checkout credentials should not be persisted in the working
tree.

The workflow track list is configured in `.github/workflows/update-patch-notices.yaml`
and maps supported snap and charm tracks to their release notes files.

## Documentation Changes

Add documentation for the `patch-notices` tool under `docs/tools/patch-notices/`,
including setup, manual usage, CI usage, metadata handling, and security notes.

## Testing

The tool should include focused unit tests for:

- release delta fetching and PR number extraction;
- workbook rendering, parsing, and clean export;
- release note insertion;
- PR body generation;
- sanitising untrusted commit, PR, and AI-generated text before rendering
  Markdown.

The GitHub Actions workflow should be validated as YAML and reviewed for secret
scoping, pinned actions, and credential persistence.

## Considerations for backwards compatibility

There are no backwards-incompatible product changes. Existing release notes
remain valid Markdown, and the metadata file is only used by this tool to decide
where the next patch notice run starts.

If the automation fails or produces unsuitable output, maintainers can continue
to update release notes manually and adjust `patch-metadata.json` in the same
PR.

## Implementation notes and guidelines

No additional implementation guidance is required beyond the sections above.