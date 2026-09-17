# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""Assemble and parse the Markdown workbook.

Workbook structure
------------------
## Included
<!-- sha:abc1234 -->
- **Major Feature** Adds XYZ support, reducing manual steps by ...

## Verification
- #42 — Original PR title

## Discarded
- #99 — Original title | _Reason: internal refactor, no user impact_
"""

from __future__ import annotations

import json
import os
import re
import tempfile
from pathlib import Path
from typing import Any

from patch_notices import sanitize

SHA_TAG_RE = re.compile(r"<!--\s*sha:([0-9a-f]{7,40})\s*-->")


def _safe_short_sha(value: Any) -> str:
    """Return a short SHA, or 'unknown' if the stored value is malformed."""
    try:
        return sanitize.short_sha(value)
    except ValueError:
        return "unknown"

CATEGORY_ORDER = [
    "Major Feature",
    "Security",
    "Deprecation",
    "Bug Fix",
    "Performance",
    "Component Bump",
    "Documentation",
]


# ---------------------------------------------------------------------------
# Rendering helpers
# ---------------------------------------------------------------------------


def _group_included(included: list[dict[str, Any]]) -> list[list[dict[str, Any]]]:
    """Return a list of groups. Standalone commits are groups of size 1.
    Commits sharing a non-null group_hint are collected together, in first-seen order.
    """
    groups: list[list[dict[str, Any]]] = []
    hint_index: dict[str, int] = {}  # group_hint -> index in groups
    for r in included:
        hint = r["triage"].get("group_hint") or None
        if hint is None:
            groups.append([r])
        elif hint in hint_index:
            groups[hint_index[hint]].append(r)
        else:
            hint_index[hint] = len(groups)
            groups.append([r])
    return groups


_CHANNEL_TAG_RE = re.compile(r"<!-- patch-notices:channel:(.+?) -->")


def _best_in_group(group: list[dict[str, Any]]) -> dict[str, Any]:
    """Return the commit from the group with the highest-priority category."""
    return min(
        group,
        key=lambda r: (
            CATEGORY_ORDER.index(r["triage"].get("category", ""))
            if r["triage"].get("category") in CATEGORY_ORDER
            else len(CATEGORY_ORDER)
        ),
    )


def _sort_by_category(items: list[dict[str, Any]]) -> list[dict[str, Any]]:
    def _key(r: dict[str, Any]) -> int:
        cat = r["triage"].get("category", "")
        try:
            return CATEGORY_ORDER.index(cat)
        except ValueError:
            return len(CATEGORY_ORDER)

    return sorted(items, key=_key)


def _components(value: Any) -> list[str]:
    return [component for item in (value or []) if (component := sanitize.component_text(item))]


def _dedupe_components(components: list[str]) -> list[str]:
    """Collapse repeat bumps of the same component to just the latest version.

    Keeps each component's first-seen position but its last-seen version, since
    *components* is built in chronological order.
    """
    latest: dict[str, str] = {}
    order: list[str] = []
    for component in components:
        name, _, _version = component.partition(" ")
        key = name.lower()
        if key not in latest:
            order.append(key)
        latest[key] = component
    return [latest[key] for key in order]


def _pr_suffix(value: Any, *, parens: bool = True) -> str:
    number = sanitize.pr_number(value)
    if number is None:
        return ""
    return f" (PR #{number})" if parens else f" PR #{number}"


def _clean_entry_lines(triage_results: list[dict[str, Any]]) -> list[str]:
    """Return clean bullet lines from triage results (no sha tags, no category labels).

    Produces the same output as export_clean but directly from triage results,
    without needing a workbook file on disk.
    """
    included = [r for r in triage_results if r["triage"]["action"] == "include"]
    groups = _group_included(_sort_by_category(included))
    lines: list[str] = []
    for group in groups:
        if len(group) == 1:
            r = group[0]
            category = sanitize.category(r["triage"].get("category", ""))
            summary = sanitize.markdown_text(r["triage"].get("summary", ""))
            components = _dedupe_components(_components(r["triage"].get("components")))
            if category == "Component Bump" and components:
                lines.append("- Version bumps")
                for component in components:
                    lines.append(f"    - {component}")
            else:
                lines.append(f"- {summary}")
        else:
            best = _best_in_group(group)
            category = sanitize.category(best["triage"].get("category", ""))
            summary = sanitize.markdown_text(best["triage"].get("summary", ""))
            all_components: list[str] = []
            for r in group:
                all_components.extend(_components(r["triage"].get("components")))
            all_components = _dedupe_components(all_components)
            if category == "Component Bump" and all_components:
                lines.append("- Version bumps")
                for component in all_components:
                    lines.append(f"    - {component}")
            else:
                lines.append(f"- {summary}")
    return lines


# ---------------------------------------------------------------------------
# Manual workflow
# ---------------------------------------------------------------------------


def write(triage_results: list[dict[str, Any]], output_path: str, channel_key: str = "") -> None:
    """Render triage results into the three-section workbook."""
    included = [r for r in triage_results if r["triage"]["action"] == "include"]
    discarded = [r for r in triage_results if r["triage"]["action"] == "discard"]

    lines: list[str] = []
    if channel_key:
        lines.append(f"<!-- patch-notices:channel:{channel_key} -->")
    lines.append("# Monthly Patch Notice Review\n")

    # -- Included ----------------------------------------------------------
    lines.append("## Included\n")
    lines.append(
        "> Edit summaries freely. Keep the `<!-- sha:... -->` tags — "
        "`finalize` uses them to update state.\n"
    )
    groups = _group_included(_sort_by_category(included))
    for group in groups:
        if len(group) == 1:
            r = group[0]
            sha = sanitize.sha(r.get("sha", ""))
            category = sanitize.category(r["triage"].get("category", ""))
            summary = sanitize.markdown_text(r["triage"].get("summary", ""))
            components = _dedupe_components(_components(r["triage"].get("components")))
            lines.append(f"<!-- sha:{sha} -->")
            if category == "Component Bump" and components:
                lines.append(f"- **{category}** Version bumps")
                for component in components:
                    lines.append(f"    - {component}")
            else:
                lines.append(f"- **{category}** {summary}")
            if r["triage"].get("limited_context"):
                lines.append("> \u26a0\ufe0f Large diff \u2014 triaged from PR title and description only. Verify before publishing.")
        else:
            # Multi-commit group: one sha tag per commit, one combined bullet
            best = _best_in_group(group)
            category = sanitize.category(best["triage"].get("category", ""))
            hint = sanitize.label_text(best["triage"].get("group_hint", ""))
            for r in group:
                lines.append(f"<!-- sha:{sanitize.sha(r.get('sha', ''))} -->")
            # Merge components from all items in the group for Component Bump groups
            all_components = []
            for r in group:
                all_components.extend(_components(r["triage"].get("components")))
            all_components = _dedupe_components(all_components)
            if category == "Component Bump" and all_components:
                lines.append(f"- **{category}** Version bumps")
                for component in all_components:
                    lines.append(f"    - {component}")
            else:
                summary = sanitize.markdown_text(best["triage"].get("summary", ""))
                lines.append(f"- **{category}** {summary}")
            covers = ", ".join(
                f"`{sanitize.short_sha(r.get('sha', ''))}` {sanitize.markdown_text(r.get('title', ''))}"
                for r in group
            )
            lines.append(f"  _(Group: {hint} — covers: {covers})_")
            if any(r["triage"].get("limited_context") for r in group):
                lines.append("> ⚠️ Large diff on one or more commits — partially triaged from PR title and description only. Verify before publishing.")
    lines.append("")

    # -- Verification ------------------------------------------------------
    lines.append("## Verification\n")
    lines.append("> Cross-check summaries against original commit titles.\n")
    for r in included:
        sha = sanitize.short_sha(r.get("sha", ""))
        pr = _pr_suffix(r.get("pr_number"))
        url = sanitize.github_url(r.get("pr_url") or r.get("html_url", ""))
        title = sanitize.markdown_text(r.get("title", ""))
        if url:
            lines.append(f"- [`{sha}`]({url}){pr} — {title}")
        else:
            lines.append(f"- `{sha}`{pr} — {title}")
    lines.append("")

    # -- Discarded ---------------------------------------------------------
    lines.append("## Discarded\n")
    lines.append(
        "> Items the AI considers noise. Move to Included (with a sha tag) "
        "if you disagree.\n"
    )
    for r in discarded:
        reason = sanitize.markdown_text(r["triage"].get("reason", ""))
        sha = sanitize.short_sha(r.get("sha", ""))
        pr = _pr_suffix(r.get("pr_number"), parens=False)
        title = sanitize.markdown_text(r.get("title", ""))
        lc_marker = " *(\u26a0\ufe0f large diff \u2014 verify)*" if r["triage"].get("limited_context") else ""
        lines.append(f"- `{sha}`{pr} \u2014 {title} | _{reason}_{lc_marker}")
    lines.append("")

    Path(output_path).write_text("\n".join(lines))


def parse_included_shas(workbook_path: str) -> list[str]:
    """Return all SHAs found in <!-- sha:... --> tags, in document order."""
    text = Path(workbook_path).read_text()
    # Only parse SHAs that appear before the Verification section
    included_section = text.split("## Verification")[0]
    return SHA_TAG_RE.findall(included_section)


def read_channel_key(workbook_path: str) -> str:
    """Return the channel key embedded by write(), e.g. 'snap:1.32-classic/stable'."""
    text = Path(workbook_path).read_text()
    m = _CHANNEL_TAG_RE.search(text)
    if not m:
        raise ValueError(
            f"No channel tag found in {workbook_path}. "
            "Re-run `review` to regenerate the workbook."
        )
    return m.group(1)


def export_clean(workbook_path: str, export_path: str) -> None:
    """Write a clean copy of the Included section, stripped of all tags."""
    text = Path(workbook_path).read_text()
    if "## Verification" not in text:
        raise ValueError(
            f"'## Verification' section not found in {workbook_path}. "
            "The workbook may be corrupted — re-run `review` to regenerate it."
        )
    included_section = text.split("## Verification")[0]
    # Remove sha tags
    clean = SHA_TAG_RE.sub("", included_section)
    # Remove blockquote editor hints
    clean = re.sub(r"^> .*\n", "", clean, flags=re.MULTILINE)
    # Remove group hint lines (indented _(Group: ...)_ lines)
    clean = re.sub(r"^ {2}_\(Group:.*\)_\n", "", clean, flags=re.MULTILINE)
    # Strip **Category** labels from bullet lines
    clean = re.sub(r"^(- )\*\*[^*]+\*\* ", r"\1", clean, flags=re.MULTILINE)
    # Collapse multiple blank lines
    clean = re.sub(r"\n{3,}", "\n\n", clean).strip() + "\n"
    Path(export_path).write_text(clean)


# ---------------------------------------------------------------------------
# CI workflow
# ---------------------------------------------------------------------------


def insert_patch_notice(
    release_notes_path: str,
    triage_results: list[dict[str, Any]],
    date: str,
    source: str,
) -> None:
    """Insert a new dated patch notice entry into the release notes file.

    If ``## Patch notices`` already exists, the new entry is prepended at the
    top of that section (newest-first order).  If the section is absent it is
    created at the canonical location:

    - snap: after ``## Upgrade notes``
    - charm: after ``## Also in this release``

    Writes the file atomically via a sibling temp file.
    """
    entry_lines = _clean_entry_lines(triage_results)
    if not entry_lines:
        return

    # Build dated entry block (no trailing blank line — caller adds separation)
    entry_block = date + "\n\n" + "\n".join(entry_lines)

    path = Path(release_notes_path)
    text = path.read_text()

    if "## Patch notices" in text:
        # Insert new entry immediately after the section heading 
        # Pattern: the heading followed by one or more blank lines.
        # Callable replacement avoids re-interpreting backslashes in entry_block as escapes.
        new_text, n = re.subn(
            r"(## Patch notices\n(?:[ \t]*\n)+)",
            lambda match: match.group(1) + entry_block + "\n\n",
            text,
            count=1,
        )
        if n == 0:
            raise ValueError(
                f"Found '## Patch notices' in {release_notes_path} but could not "
                "match expected heading format '## Patch notices\\n\\n'. "
                "Check for unexpected whitespace."
            )
    else:
        # Section absent — create it after the anchor heading's content block.
        anchor = "## Upgrade notes" if source == "snap" else "## Also in this release"
        if anchor not in text:
            raise ValueError(
                f"Cannot find anchor heading '{anchor}' in {release_notes_path}. "
                "Unable to create '## Patch notices' section."
            )
        anchor_pos = text.index(anchor)
        # Walk past the anchor to find the next ## heading
        after_anchor = text[anchor_pos + len(anchor):]
        next_heading = re.search(r"\n##\s", after_anchor)
        new_section = f"\n## Patch notices\n\n{entry_block}\n"
        if next_heading:
            insert_pos = anchor_pos + len(anchor) + next_heading.start()
            new_text = text[:insert_pos] + new_section + text[insert_pos:]
        else:
            new_text = text.rstrip() + new_section + "\n"

    # Atomic write: write to temp then rename
    fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=".tmp-patch-notice-")
    try:
        with os.fdopen(fd, "wb") as fh:
            fh.write(new_text.encode())
        os.replace(tmp, path)
    except Exception:
        os.unlink(tmp)
        raise


def build_track_summary(
    triage_results: list[dict[str, Any]],
    track: str,
    date: str,
) -> dict[str, Any]:
    """Return a per-track summary dict for the pr-body command.

    *track* is the state key, e.g. ``snap:1.35-classic/stable``.
    """
    included = [r for r in triage_results if r["triage"]["action"] == "include"]
    discarded = [r for r in triage_results if r["triage"]["action"] == "discard"]
    # Low-context flag applies regardless of action, so discards need the warning too.
    limited = [r for r in triage_results if r["triage"].get("limited_context")]

    return {
        "track": track,
        "status": "updated" if included else "all-discarded",
        "date": date,
        "included": [
            {
                "sha": r.get("sha", "")[:8],
                "pr_number": r.get("pr_number"),
                "title": r.get("title", ""),
                "category": r["triage"].get("category", ""),
                "summary": r["triage"].get("summary", ""),
                "components": r["triage"].get("components"),
            }
            for r in included
        ],
        "discarded": [
            {
                "sha": r.get("sha", "")[:8],
                "pr_number": r.get("pr_number"),
                "title": r.get("title", ""),
                "reason": r["triage"].get("reason", ""),
            }
            for r in discarded
        ],
        "limited_context": [
            {
                "sha": r.get("sha", "")[:8],
                "pr_number": r.get("pr_number"),
                "title": r.get("title", ""),
            }
            for r in limited
        ],
    }


def build_pr_body(summaries: list[dict[str, Any]]) -> str:
    """Generate GitHub PR body markdown from a list of per-track summary dicts."""
    from datetime import date as _date

    today = _date.today().strftime("%Y-%m-%d")

    def _parse_track(track: str) -> tuple[str, str]:
        """Return (source, human-readable version) from a state key."""
        source, rest = track.split(":", 1)
        version = rest.split("/")[0].replace("-classic", "")
        return sanitize.markdown_text(source), sanitize.markdown_text(version)

    # Sort: snap before charm, newest version first within each source
    def _sort_key(s: dict[str, Any]) -> tuple[int, str]:
        src, ver = _parse_track(s["track"])
        return (0 if src == "snap" else 1, ver)

    summaries = sorted(summaries, key=_sort_key, reverse=False)
    # reverse version within source so newest is first
    snap = sorted([s for s in summaries if s["track"].startswith("snap:")],
                  key=lambda s: _parse_track(s["track"])[1], reverse=True)
    charm = sorted([s for s in summaries if s["track"].startswith("charm:")],
                   key=lambda s: _parse_track(s["track"])[1], reverse=True)
    summaries = snap + charm

    lines: list[str] = [f"## Patch notices — {today}\n"]

    # Summary table
    lines.append("| Release | Source | Status | Included | Discarded | ⚠️ |")
    lines.append("|---|---|---|---|---|---|")
    for s in summaries:
        source, version = _parse_track(s["track"])
        if s["status"] == "updated":
            status_str = "✅ Updated"
        elif s["status"] == "up-to-date":
            status_str = "— Up to date"
        else:
            status_str = "— All discarded"
        n_inc = len(s.get("included", []))
        n_dis = len(s.get("discarded", []))
        n_lc = len(s.get("limited_context", []))
        lines.append(
            f"| {version} | {source} | {status_str} "
            f"| {n_inc or ''} | {n_dis or ''} | {n_lc or ''} |"
        )
    lines.append("")

    # Collapsible detail blocks — only for tracks with activity
    for s in summaries:
        if s["status"] == "up-to-date":
            continue
        source, version = _parse_track(s["track"])
        n_inc = len(s.get("included", []))
        n_dis = len(s.get("discarded", []))
        n_lc = len(s.get("limited_context", []))
        label = f"{source.capitalize()} {version}"
        parts = []
        if n_inc:
            parts.append(f"{n_inc} included")
        if n_dis:
            parts.append(f"{n_dis} discarded")
        if n_lc:
            parts.append(f"{n_lc} ⚠️")
        if parts:
            label += " — " + ", ".join(parts)

        lines.append(f"<details><summary>{label}</summary>")
        lines.append("")

        if s.get("included"):
            lines.append("### Included")
            lines.append("")
            for item in s["included"]:
                sha = _safe_short_sha(item.get("sha", ""))
                pr = _pr_suffix(item.get("pr_number"))
                title = sanitize.markdown_text(item.get("title", ""))
                lines.append(f"- `{sha}`{pr} — {title}")
            lines.append("")

        if s.get("limited_context"):
            lines.append("### ⚠️ Large diff — verify manually before approving")
            lines.append("")
            for item in s["limited_context"]:
                sha = _safe_short_sha(item.get("sha", ""))
                pr = _pr_suffix(item.get("pr_number"))
                title = sanitize.markdown_text(item.get("title", ""))
                lines.append(f"- `{sha}`{pr} — {title}")
                lines.append("  _(triaged from PR title and description only)_")
            lines.append("")

        if s.get("discarded"):
            lines.append("### Discarded")
            lines.append("")
            for item in s["discarded"]:
                sha = _safe_short_sha(item.get("sha", ""))
                pr = _pr_suffix(item.get("pr_number"), parens=False)
                reason = sanitize.markdown_text(item.get("reason", ""))
                title = sanitize.markdown_text(item.get("title", ""))
                lines.append(f"- `{sha}`{pr} — {title}")
                if reason:
                    lines.append(f"  _{reason}_")
            lines.append("")

        lines.append("</details>")
        lines.append("")

    return "\n".join(lines)