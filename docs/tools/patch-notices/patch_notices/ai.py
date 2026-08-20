# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""AI triage: categorise and summarise each PR in the delta.

Truth hierarchy (per spec): file diffs > PR body > PR title.
"""

from __future__ import annotations

import json
import os
from typing import Any

import openai

SNAP_SYSTEM_PROMPT = """\
You are a technical writer producing monthly patch notices for Canonical Kubernetes.
You will be given a git commit: its title, body, and the file diff for that commit.

## Editorial strategy

- **Focus on impact.** Do not repeat the commit subject. Describe the user-facing
  benefit, operational impact, or reason the change matters.
- **Truth hierarchy.** File diffs > PR body > PR title. Ignore misleading titles when
  the diff tells a different story.
- **Filter for user action.** Include: bug fixes, features, security-relevant
  dependency updates, deprecation warnings, operational improvements, and significant
  upgrade or rollback behaviour changes.
- **Major docs only.** Include documentation changes only when they introduce a new
  guide, a new supported workflow, a deprecation notice, or a material change in how
  an operator uses or manages the product.
- **Exclude noise.** Discard: copyright changes, linting-only changes, CI-only
  changes, test-only changes, release preparation, and other low-signal maintenance
  unless they directly affect shipped behaviour.
- **Patch notice and release note updates.** Discard commits whose sole purpose
  is updating, backporting, or amending patch notices, release notes, or changelogs
  (e.g. "Update release notes patch notices", "docs: backport patch notice update").
  These are meta-documentation — the changes they describe were already captured by
  earlier commits in the delta.
- **Strict revision exclusion.** Discard commits that only update
  architecture-specific snap revision numbers, e.g.
  `Update K8s revisions ["amd64-xxxx", "arm64-xxxx"]`.

## Snap-specific rules

- **Include `Update component versions` commits.** These represent real shipped
  changes and must never be excluded.
- **Distinguish component bumps from revision bookkeeping:**
  - *Include*: packaged component updates — Kubernetes, CNI, containerd, runc, Helm,
    k8s-dqlite, Cilium, CoreDNS, MetalLB, metrics-server, and similar shipped deps.
  - *Exclude*: commits that only move track-specific amd64/arm64 snap revision numbers.
- **Version bumps formatting.** When component updates are present, populate the
  `components` array (see Output format). List each updated component as
  `"name new-version"`, e.g. `"containerd v1.7.30"`.
  Only include components where the new version is visible in the diff.
  If versions are not visible, set `components` to null and use `summary` instead.

## Tone

Professional, concise, and focused on stability, upgrade safety, security, and
operator experience.

## Output format

Respond with valid JSON only — no markdown fences:
{
  "action": "include" | "discard",
  "category": "Major Feature" | "Deprecation" | "Bug Fix" | "Security" |
              "Component Bump" | "Performance" | "Documentation" | null,
  "summary": "<starts with a verb, states the fix/feature AND its consequence for the operator, max 120 chars, or null>",
  // Prefer: 'Honor AnnotationDisableSeparateFeatureUpgrades during node joins to prevent unintended upgrades.'
  // Over:   'Honor AnnotationDisableSeparateFeatureUpgrades during node joins.'
  "components": ["<name> <new-version>", ...] | null,
  // Only for Component Bump when the new version is visible. e.g. ["containerd v1.7.30", "runc v1.3.4"]
  // null for all other categories and for Component Bumps where versions are not in the diff.
  "reason": "<one sentence reason if discarded, else null>"
}
"""


CHARM_SYSTEM_PROMPT = """\
You are a technical writer producing monthly patch notices for Canonical Kubernetes charms.
You will be given a git commit from the k8s-operator repository: its title, body, and the
file diff for that commit.

## Editorial strategy

- **Focus on operator impact.** Describe what changes for a Juju operator managing a
  Canonical Kubernetes cluster — new options, changed behaviour, fixed bugs, or new
  capabilities.
- **Truth hierarchy**: File diffs > PR body > PR title. Ignore misleading titles when
  the diff tells a different story.
- **Categories**: Major Features, Deprecations, Bug Fixes, Security,
  Component Bumps, Performance, Documentation, Internal (discarded).

## Include

- **Config option changes**: new options added to `charmcraft.yaml` config, option
  defaults changed, options removed or deprecated.
- **Action changes**: new Juju actions, changed action parameters or output, removed actions.
- **Relation changes**: new integration endpoints added or removed, changed relation
  interfaces that affect how the charm integrates with other charms.
- **OCI/rock image bumps**: when the charm ships a new OCI image or rock version that
  changes operator-visible behaviour or fixes bugs.
- **Hook and event handler bug fixes**: fixes in charm hooks (`install`, `upgrade-charm`,
  `config-changed`, relation hooks, etc.) that change observable cluster behaviour.
- **Bootstrap and upgrade behaviour changes**: anything that changes how `k8s bootstrap`,
  cluster join, or charm upgrade works from an operator's perspective.
- **User-facing documentation changes**: changes to `charmcraft.yaml` descriptions,
  README, or operator guides that materially change how an operator uses the product.

## Discard

- **Juju ops/libs library bumps**: updating `ops`, `cosl`, or other Juju library
  dependencies with no visible behaviour change (e.g. `Bump ops to 2.x`).
- **snap-installation resource revision bumps**: commits that only update which snap
  revision the charm installs — these are covered by the snap patch notices.
- **CI-only changes**: GitHub workflow files, tox configs, Makefile targets, and similar
  that do not affect the shipped charm.
- **Linting and code style**: `ruff`, `black`, `isort`, `mypy` fixes with no logic change.
- **Test-only changes**: unit tests, integration tests, spread tests, fixtures — nothing
  ships to operators.
- **Release preparation**: version bumps, CHANGELOG updates, release commit messages.
- **Patch notice and release note updates**: commits that update, backport, or amend the
  patch notices or release notes themselves — these are meta-documentation, not new
  operator-facing changes.
- **Copyright and license header changes**: legal boilerplate with no functional change.
- **Internal refactors**: code restructuring with no operator-visible behaviour change.

## Tone

Professional, concise, and focused on stability, upgrade safety, security, and
operator experience. Start summaries with a verb (e.g. "Add", "Fix", "Remove").

## Output format

Respond with valid JSON only — no markdown fences:
{
  "action": "include" | "discard",
  "category": "Major Feature" | "Deprecation" | "Bug Fix" | "Security" |
              "Component Bump" | "Performance" | "Documentation" | null,
  "summary": "<starts with a verb, states the fix/feature AND its consequence for the operator, max 120 chars, or null>",
  // Prefer: 'Honor AnnotationDisableSeparateFeatureUpgrades during node joins to prevent unintended upgrades.'
  // Over:   'Honor AnnotationDisableSeparateFeatureUpgrades during node joins.'
  "components": null,
  // Always null for charm triage (charm releases do not use the Version bumps nested format).
  "reason": "<one sentence reason if discarded, else null>"
}
"""


GROUP_PASS_PROMPT = """\
You are grouping a list of already-triaged commits for a Canonical Kubernetes patch notice.

Each commit has been individually reviewed and approved for inclusion. Your task is to
identify which commits cover the same feature story, bug fix, or related work so they
can be rendered as a single combined entry in the release notes.

## Rules

- Assign a short label (e.g. "CoreDNS HA", "BGP relation", "upgrade hook fix") to
  commits that belong to the same logical story.
- Use null for standalone commits that do not share a story with any other commit.
- Be conservative: only group commits when the connection is clear and obvious.
  When in doubt, leave commits as standalone (null).
- A group must have at least 2 members — do not create a group for a single commit.

## Output format

Respond with valid JSON only — no markdown fences.
Return an object where each key is the full commit SHA from the input and the value
is a group label string (if grouped) or null (if standalone):
{
  "<full-sha>": "<short group label>" | null,
  ...
}
"""

# Character cap for the diff included in each triage prompt.
# When a commit's diff exceeds this limit the diff is omitted entirely and the
# commit is flagged in the workbook for extra reviewer attention.
# Override via PATCH_NOTICES_MAX_DIFF_CHARS if your endpoint has a tighter limit.
_MAX_DIFF_CHARS = int(os.environ.get("PATCH_NOTICES_MAX_DIFF_CHARS", "12000"))


def triage(prs: list[dict[str, Any]], source: str = "snap") -> list[dict[str, Any]]:
    """Run each commit through two LLM passes. Returns an enriched list of records.

    Pass 1 — individual triage: each commit is evaluated in isolation using its own
    diff (include/discard, category, summary). Commits whose diff exceeds
    _MAX_DIFF_CHARS have the diff omitted and are flagged with limited_context=True.

    Pass 2 — grouping: a single call assigns group_hints to included commits,
    identifying related commits to render as one combined workbook entry.

    Supports OpenAI directly or any OpenAI-compatible endpoint.
    Set OPENAI_BASE_URL to override, e.g.:
      export OPENAI_BASE_URL=https://models.inference.ai.azure.com
      export OPENAI_API_KEY=ghp_...
    """
    system_prompt = CHARM_SYSTEM_PROMPT if source == "charm" else SNAP_SYSTEM_PROMPT
    client = openai.OpenAI(
        api_key=os.environ["OPENAI_API_KEY"],
        base_url=os.environ.get("OPENAI_BASE_URL") or None,  # empty string → None
    )

    # Pass 1: triage each commit individually
    results = []
    for pr in prs:
        result = _triage_one(client, pr, system_prompt)
        results.append({**pr, "triage": result})

    # Pass 2: assign group_hints to included commits (needs ≥2 to form any group)
    included = [r for r in results if r["triage"]["action"] == "include"]
    if len(included) >= 2:
        group_map = _group_pass(client, included)
        for r in results:
            if r["triage"]["action"] == "include":
                r["triage"]["group_hint"] = group_map.get(r["sha"])

    return results


def _group_pass(client: openai.OpenAI, included: list[dict[str, Any]]) -> dict[str, str | None]:
    """Second pass: assign group_hints to included commits via a single LLM call.

    Uses OPENAI_GROUP_MODEL (default: gpt-4o-mini) — grouping from plain-English
    summaries is simpler than diff triage and does not need a large model.

    Falls back gracefully to an empty map on any error so the workbook is still
    produced without groups rather than failing the run.
    """
    items = [
        {
            "sha": r["sha"],
            "title": r.get("title", ""),
            "summary": r["triage"].get("summary", ""),
            "category": r["triage"].get("category", ""),
        }
        for r in included
    ]

    try:
        response = client.chat.completions.create(
            model=os.environ.get("OPENAI_GROUP_MODEL", "gpt-4o-mini"),
            response_format={"type": "json_object"},
            messages=[
                {"role": "system", "content": GROUP_PASS_PROMPT},
                {"role": "user", "content": json.dumps(items, indent=2)},
            ],
            temperature=0.1,
        )
        return json.loads(response.choices[0].message.content)
    except Exception as exc:
        print(f"[warning] Group pass failed ({exc}); group_hints will be null.", flush=True)
        return {}


def _triage_one(client: openai.OpenAI, pr: dict[str, Any], system_prompt: str) -> dict[str, Any]:
    """Triage a single commit. Returns the parsed JSON response.

    If the commit's diff exceeds _MAX_DIFF_CHARS it is omitted and
    limited_context is set True so the workbook can flag the entry for
    extra reviewer attention.
    """
    raw_diff = pr.get("diff") or ""
    if raw_diff and len(raw_diff) > _MAX_DIFF_CHARS:
        diff_content = "(diff omitted — too large; triaged from PR title and body only)"
        limited_context = True
    else:
        diff_content = raw_diff or "(not available)"
        limited_context = False

    user_content = (
        f"Commit {pr.get('sha', '')[:8]}"
        + (f" (PR #{pr.get('pr_number')})" if pr.get('pr_number') else "")
        + f": {pr.get('title')}\n\n"
        f"Author: {pr.get('author')}\n"
        f"Date: {pr.get('date')}\n\n"
        f"Body:\n{pr.get('body') or '(none)'}\n\n"
        f"Diff:\n{diff_content}"
    )
    response = client.chat.completions.create(
        model=os.environ.get("OPENAI_MODEL", "gpt-4o"),
        response_format={"type": "json_object"},
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_content},
        ],
        temperature=0.2,
    )
    try:
        result = json.loads(response.choices[0].message.content)
    except Exception as exc:
        sha = pr.get("sha", "")[:8]
        print(f"[warning] JSON parse failed for {sha} ({exc}); flagging for manual review.", flush=True)
        result = {
            "action": "include",
            "category": None,
            "summary": "(AI response unparseable — review manually)",
            "components": None,
            "reason": None,
        }
    result["group_hint"] = None  # populated by _group_pass if applicable
    result["limited_context"] = limited_context
    result.setdefault("components", None)  # ensure field exists if AI omits it
    return result