# docs

MkDocs-based user documentation (`canonicalk8s/`) and design proposals (`proposals/`).

## Design Proposals

Non-trivial features require a proposal before implementation. Proposals live in
`docs/proposals/` and are numbered sequentially.

To create one:

1. Copy `docs/proposals/000-template.md` to `docs/proposals/<NNN>-<slug>.md`
2. Fill every section (incomplete proposals should stay in `DRAFTING`)

Status values: `DRAFTING`, `ACCEPTED`, `REJECTED`.

### Required Sections

| Section | Requirement |
|---------|-------------|
| Summary | short paragraph; what and why |
| Rationale | user scenarios, problem being solved |
| User facing changes | all CLI/API/output changes with before/after examples |
| Alternative solutions | options considered and why rejected |
| Out of scope | explicit exclusions and unknowns |
| API Changes | new/modified k8sd endpoints and message types |
| CLI Changes | new arguments, changed output formats |
| Database Changes | schema additions or migrations |
| Configuration Changes | new service args or config fields |
| Documentation Changes | pages to add or update |
| Testing | how the feature will be tested |
| Backwards compatibility | breaking changes and how older clients are handled |
| Implementation notes | code pointers and guidance for implementors |

Link to specific commits (not branches) when referencing code in proposals so links
remain valid after the branch is deleted.

## Spread Tests

`docs/spread.yaml` drives doc-embedded tests via [Spread](https://github.com/snapcore/spread).
Two suites:

- `tests/spread-generated/snap_clean/` — installs and bootstraps the snap from scratch
- `tests/spread-generated/snap_bootstrapped/` — assumes a running cluster

These run in CI against `ubuntu-24.04` via the `github-ci` adhoc backend.

## Documentation Build

```
cd docs/canonicalk8s
make html       # build locally (requires Python deps from requirements.txt)
```
