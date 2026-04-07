#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Dependency update subcommands for ``k8s-ci``.

Provides tooling to keep CI dependency pins (GitHub Actions SHA pins and
pip hash-pinned lock files) up to date.
"""

import argparse
import glob
import json
import os
import re
import subprocess
import sys
import uuid
from pathlib import Path
from typing import Optional

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
    p.set_defaults(func=cmd_update_pip_locks)


# -- GitHub API helpers ------------------------------------------------------


def _gh_api(endpoint: str) -> Optional[dict | list]:
    """Call the GitHub REST API via ``gh api``.  Returns parsed JSON or None."""
    try:
        r = subprocess.run(
            ["gh", "api", endpoint, "--paginate"],
            capture_output=True,
            text=True,
            check=True,
            timeout=30,
        )
        return json.loads(r.stdout)
    except (
        subprocess.CalledProcessError,
        json.JSONDecodeError,
        subprocess.TimeoutExpired,
    ) as e:
        print(f"  WARNING: gh api {endpoint}: {e}", file=sys.stderr)
        return None


def _resolve_sha(owner_repo: str, ref: str) -> Optional[str]:
    """Resolve a tag or branch to its commit SHA (dereferences annotated tags)."""
    data = _gh_api(f"repos/{owner_repo}/git/ref/tags/{ref}")
    if data is None:
        data = _gh_api(f"repos/{owner_repo}/git/ref/heads/{ref}")
    if not isinstance(data, dict):
        return None

    obj = data.get("object", {})
    sha, kind = obj.get("sha"), obj.get("type")
    if kind == "tag" and sha:
        tag_obj = _gh_api(f"repos/{owner_repo}/git/tags/{sha}")
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

    data = _gh_api(f"repos/{owner_repo}/git/matching-refs/tags/v{major}")
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


def cmd_update_pip_locks(args: argparse.Namespace) -> int:
    """Recompile pip lock files and report changes.

    Args:
        args: Parsed command-line arguments.

    Returns:
        0 on success, 1 on compile failure.
    """
    root = Path(args.root).resolve()
    in_files = sorted(glob.glob(str(root / "ci" / "requirements-*.in")))

    if not in_files:
        print("No ci/requirements-*.in files found.")
        _set_output("has-changes", "false")
        return 0

    changed_files = []
    for infile in in_files:
        outfile = infile.removesuffix(".in") + ".txt"
        print(f"Compiling {Path(infile).relative_to(root)} ...")
        try:
            r = subprocess.run(
                [
                    "pip-compile",
                    "--upgrade",
                    "--generate-hashes",
                    "--strip-extras",
                    "--no-header",
                    "--output-file",
                    outfile,
                    infile,
                ],
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
                f"ERROR: pip-compile failed for {infile}:\n{r.stderr}", file=sys.stderr
            )
            return 1

        # Check for changes via git diff
        diff = subprocess.run(
            ["git", "diff", "--quiet", outfile],
            capture_output=True,
        )
        if diff.returncode != 0:
            changed_files.append(Path(outfile).relative_to(root))

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


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_deps_cmds(parser)
    args_main = parser.parse_args()
    sys.exit(args_main.func(args_main))
