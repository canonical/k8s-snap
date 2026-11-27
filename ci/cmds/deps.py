#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Dependency update subcommands for ``k8s-ci``.

Provides tooling to keep CI dependency pins (GitHub Actions SHA pins and
pip hash-pinned lock files) up to date.
"""

from __future__ import annotations

import argparse
import glob
import os
import re
import subprocess
import sys
import tempfile
import uuid
from pathlib import Path
from typing import Dict, List, NamedTuple, Optional

from cmds.github import gh_api

_REPO_ROOT = Path(__file__).resolve().parents[2]

# owner/repo@<40-hex-sha> # <tag>
_USES_RE = re.compile(
    r"(uses:\s+)"
    r"([a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+)"
    r"@([0-9a-f]{40})"
    r"\s+#\s*(\S+)"
)

_SEMVER_RE = re.compile(r"^v?(\d+)(?:\.(\d+)(?:\.(\d+))?)?$")

_FLOATING = {"main", "master", "latest"}

_STANZA_START_RE = re.compile(r"^([a-zA-Z0-9_.-]+)(?:\[[^\]]+\])?==([^\s\\;#]+)")


class _RequirementStanza(NamedTuple):
    name: str
    version: object
    lines: List[str]


# -- CLI registration -------------------------------------------------------


def add_deps_cmds(parser):
    """Register dependency-related subcommands."""
    deps = parser.add_parser("deps", help="CI dependency management commands.")
    sub = deps.add_subparsers(dest="deps_command", required=True, title="deps commands")

    p = sub.add_parser(
        "update-actions", help="Update SHA-pinned GitHub Actions to latest versions."
    )
    p.add_argument("--root", default=str(_REPO_ROOT), help="Repository root directory.")
    p.set_defaults(func=cmd_update_actions)

    p = sub.add_parser(
        "update-pip-locks", help="Recompile pip lock files with latest versions."
    )
    p.add_argument("--root", default=str(_REPO_ROOT), help="Repository root directory.")
    p.add_argument(
        "--file",
        dest="files",
        action="append",
        default=[],
        help="Specific requirements .in file(s) to compile (default: all ci/requirements-*.in).",
    )
    p.set_defaults(func=cmd_update_pip_locks)


def _resolve_sha(owner_repo: str, ref: str) -> Optional[str]:
    """Resolve a tag or branch to its commit SHA (dereferences annotated tags)."""
    data = gh_api(f"repos/{owner_repo}/git/ref/tags/{ref}")
    if data is None:
        data = gh_api(f"repos/{owner_repo}/git/ref/heads/{ref}")
    if not isinstance(data, dict):
        return None

    obj = data.get("object", {})
    sha, kind = obj.get("sha"), obj.get("type")
    if kind == "tag" and sha:
        tag_obj = gh_api(f"repos/{owner_repo}/git/tags/{sha}")
        if isinstance(tag_obj, dict):
            sha = tag_obj.get("object", {}).get("sha", sha)
    return sha


# -- Version helpers ---------------------------------------------------------


def _semver(tag: str) -> Optional[tuple[int, int, int]]:
    m = _SEMVER_RE.match(tag)
    return (int(m.group(1)), int(m.group(2) or 0), int(m.group(3) or 0)) if m else None


def _is_floating(tag: str) -> bool:
    if tag in _FLOATING:
        return True
    m = _SEMVER_RE.match(tag)
    return bool(m and m.group(2) is None)


def _find_latest(owner_repo: str, tag: str) -> Optional[tuple[str, str]]:
    """Return ``(new_tag, sha)`` if a newer same-major version exists."""
    cur = _semver(tag)
    if cur is None:
        return None
    major = cur[0]

    data = gh_api(f"repos/{owner_repo}/git/matching-refs/tags/v{major}")
    if not isinstance(data, list):
        return None

    best, best_tag = cur, tag
    for item in data:
        name = item.get("ref", "").removeprefix("refs/tags/")
        v = _semver(name)
        if not v or v[0] != major:
            continue
        m = _SEMVER_RE.match(name)
        if m and m.group(2) is None:
            continue  # skip floating (e.g. "v4")
        if v > best:
            best, best_tag = v, name

    if best_tag == tag:
        return None
    sha = _resolve_sha(owner_repo, best_tag)
    return (best_tag, sha) if sha else None


# -- Scan / update actions ---------------------------------------------------


def _scan_actions(root: Path) -> list[tuple[Path, int, str, str, str]]:
    """Return list of ``(file, line, action, sha, tag)`` tuples."""
    refs = []
    for d in (root / ".github" / "workflows", root / ".github" / "actions"):
        if not d.exists():
            continue
        for f in sorted(d.rglob("*.y*ml")):
            for n, line in enumerate(f.read_text().splitlines(), 1):
                m = _USES_RE.search(line)
                if m:
                    refs.append((f, n, m.group(2), m.group(3), m.group(4)))
    return refs


def _check_update(action: str, tag: str, sha: str) -> Optional[tuple[str, str, str]]:
    """Return ``(new_tag, new_sha, reason)`` or None."""
    if _is_floating(tag):
        new = _resolve_sha(action, tag)
        if new and new != sha:
            return (tag, new, f"floating tag `{tag}` moved")
        return None

    result = _find_latest(action, tag)
    if result:
        return (result[0], result[1], f"`{tag}` -> `{result[0]}`")

    new = _resolve_sha(action, tag)
    if new and new != sha:
        return (tag, new, f"tag `{tag}` SHA changed")
    return None


# -- GitHub Actions output ---------------------------------------------------


def _set_output(name: str, value: str) -> None:
    path = os.environ.get("GITHUB_OUTPUT")
    if not path:
        return
    with open(path, "a") as f:
        if "\n" in value:
            delim = f"ghadelimiter_{uuid.uuid4()}"
            f.write(f"{name}<<{delim}\n{value}\n{delim}\n")
        else:
            f.write(f"{name}={value}\n")


# -- Commands ----------------------------------------------------------------


def cmd_update_actions(args: argparse.Namespace) -> int:
    """Update SHA-pinned GitHub Actions references.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 on success.
    """
    root = Path(args.root).resolve()
    refs = _scan_actions(root)
    print(f"Found {len(refs)} SHA-pinned action reference(s).")

    if not refs:
        _set_output("has-changes", "false")
        return 0

    # Deduplicate API calls by (action, tag)
    cache: dict[tuple[str, str], Optional[tuple[str, str, str]]] = {}
    updates = []  # (file, action, old_sha, new_sha, old_tag, new_tag, reason)

    for filepath, line, action, sha, tag in refs:
        key = (action, tag)
        if key not in cache:
            cache[key] = _check_update(action, tag, sha)
        result = cache[key]
        if result and result[1] != sha:
            updates.append(
                (filepath, action, sha, result[1], tag, result[0], result[2])
            )

    if not updates:
        print("All actions are up to date.")
        _set_output("has-changes", "false")
        return 0

    # Apply updates grouped by file
    files: dict[Path, list] = {}
    for u in updates:
        files.setdefault(u[0], []).append(u)
    for filepath, file_updates in files.items():
        content = filepath.read_text()
        for _, action, old_sha, new_sha, old_tag, new_tag, _ in file_updates:
            content = content.replace(
                f"{action}@{old_sha} # {old_tag}",
                f"{action}@{new_sha} # {new_tag}",
            )
        filepath.write_text(content)

    # Summary
    seen = {}
    for _, action, _, _, _, _, reason in updates:
        seen.setdefault(action, reason)
    summary = "\n".join(f"- `{a}`: {r}" for a, r in sorted(seen.items()))
    print(f"Updated {len(updates)} reference(s) across {len(files)} file(s).")
    print(summary)

    _set_output("has-changes", "true")
    _set_output("summary", summary)
    return 0


def _parse_requirement_stanzas(content: str) -> Dict[str, _RequirementStanza]:
    """Parse requirement stanzas from locked content."""
    from packaging.utils import canonicalize_name
    from packaging.version import Version

    stanzas: Dict[str, _RequirementStanza] = {}
    curr_name: Optional[str] = None
    curr_version: Optional[Version] = None
    curr_lines: List[str] = []

    for line in content.splitlines():
        if not line.strip() or line.startswith("#"):
            continue
        m = _STANZA_START_RE.match(line)
        if m:
            if curr_name is not None and curr_version is not None:
                canonical = canonicalize_name(curr_name)
                stanzas[canonical] = _RequirementStanza(
                    curr_name, curr_version, curr_lines
                )
            curr_name = m.group(1)
            curr_version = Version(m.group(2))
            curr_lines = [line]
        elif curr_name is not None:
            curr_lines.append(line)

    if curr_name is not None and curr_version is not None:
        canonical = canonicalize_name(curr_name)
        stanzas[canonical] = _RequirementStanza(curr_name, curr_version, curr_lines)

    return stanzas


def _merge_requirement_stanzas(base_content: str, overrides_content: str) -> str:
    """Merge compiled overrides into base requirements without downgrading."""
    from packaging.utils import canonicalize_name

    base_stanzas = _parse_requirement_stanzas(base_content)
    override_stanzas = _parse_requirement_stanzas(overrides_content)

    for canonical, override in override_stanzas.items():
        if canonical not in base_stanzas:
            base_stanzas[canonical] = override
        elif override.version > base_stanzas[canonical].version:
            base_stanzas[canonical] = override

    sorted_stanzas = sorted(
        base_stanzas.values(), key=lambda s: canonicalize_name(s.name)
    )
    out_lines: List[str] = []
    for stanza in sorted_stanzas:
        out_lines.extend(stanza.lines)

    return "\n".join(out_lines) + "\n"


def cmd_update_pip_locks(args: argparse.Namespace) -> int:
    """Recompile pip lock files and report changes.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 on success, 1 on compile failure.
    """
    root = Path(args.root).resolve()
    if getattr(args, "files", None):
        in_files = []
        for f in args.files:
            p = Path(f)
            p = (p if p.is_absolute() else root / p).resolve()
            if not p.is_relative_to(root) or p.suffix != ".in" or not p.is_file():
                print(f"ERROR: Invalid requirements file: {f}", file=sys.stderr)
                return 1
            in_files.append(str(p))
        in_files.sort()
    else:
        in_files = sorted(glob.glob(str(root / "ci" / "requirements-*.in")))
    if not in_files:
        print("No ci/requirements-*.in files found.")
        _set_output("has-changes", "false")
        return 0

    changed_files = []
    for infile in in_files:
        rel_in = str(Path(infile).relative_to(root))
        rel_out = rel_in.removesuffix(".in") + ".txt"
        overrides_file = Path(infile).with_suffix(".overrides")
        has_overrides = overrides_file.is_file()

        print(f"Compiling {rel_in} ...")
        with tempfile.TemporaryDirectory() as tmp_dir:
            base_out = str(Path(tmp_dir) / "base.txt") if has_overrides else rel_out
            try:
                r = subprocess.run(
                    [
                        "pip-compile",
                        "--upgrade",
                        "--allow-unsafe",
                        "--generate-hashes",
                        "--strip-extras",
                        "--no-header",
                        "--output-file",
                        base_out,
                        rel_in,
                    ],
                    cwd=root,
                    capture_output=True,
                    text=True,
                )
            except FileNotFoundError:
                print(
                    "ERROR: pip-compile not found. Install it with: pip install pip-tools",
                    file=sys.stderr,
                )
                return 1
            if r.returncode != 0:
                print(
                    f"ERROR: pip-compile failed for {rel_in}:\n{r.stderr}",
                    file=sys.stderr,
                )
                return 1

            if has_overrides:
                rel_overrides = str(overrides_file.relative_to(root))
                print(f"Applying overrides from {rel_overrides} ...")
                overrides_out = Path(tmp_dir) / "overrides.txt"
                try:
                    r_overrides = subprocess.run(
                        [
                            "pip-compile",
                            "--upgrade",
                            "--allow-unsafe",
                            "--generate-hashes",
                            "--strip-extras",
                            "--no-header",
                            "--output-file",
                            str(overrides_out),
                            rel_overrides,
                        ],
                        cwd=root,
                        capture_output=True,
                        text=True,
                    )
                except FileNotFoundError:
                    print(
                        "ERROR: pip-compile not found. Install it with: pip install pip-tools",
                        file=sys.stderr,
                    )
                    return 1
                if r_overrides.returncode != 0:
                    print(
                        f"ERROR: pip-compile failed for {rel_overrides}:\n{r_overrides.stderr}",
                        file=sys.stderr,
                    )
                    return 1

                locked_file = root / rel_out
                merged = _merge_requirement_stanzas(
                    Path(base_out).read_text(),
                    overrides_out.read_text(),
                )
                locked_file.write_text(merged)
        # Check for modified or untracked lock files
        status = subprocess.run(
            ["git", "status", "--porcelain", "--", rel_out],
            cwd=root,
            capture_output=True,
            text=True,
        )
        if status.stdout.strip():
            changed_files.append(Path(rel_out))
    if not changed_files:
        print("All pip lock files are up to date.")
        _set_output("has-changes", "false")
        return 0

    summary = "\n".join(f"- `{f}` updated" for f in changed_files)
    print(f"Updated {len(changed_files)} lock file(s).")
    print(summary)

    _set_output("has-changes", "true")
    _set_output("summary", summary)
    return 0
