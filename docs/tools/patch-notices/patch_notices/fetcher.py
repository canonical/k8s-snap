# Copyright 2026 Canonical, Ltd.
# See LICENSE file for licensing details.

"""Fetch the commit delta between the last documented SHA and the current stable SHA.

Snap pipeline:
  1. Query Snap Store API  -> current revision number for the track
  2. Query Launchpad API   -> git SHA for that revision's build
  3. Query GitHub API      -> list of commits merged between the two SHAs

Charm pipeline:
  1. Query Charmhub API    -> current revision number for the track
  2. Query GitHub API      -> git SHA via the k8s-rev<N> tag
  3. Query GitHub API      -> list of commits merged between the two SHAs
"""

from __future__ import annotations

import json
import os
import pathlib
import re
import sys
from typing import Any

import requests

from patch_notices import metadata

SNAP_STORE_API = "https://api.snapcraft.io/v2/snaps/info/k8s"
LAUNCHPAD_API = "https://api.launchpad.net/devel"
GITHUB_API = "https://api.github.com"
GITHUB_REPO = "canonical/k8s-snap"
CHARMHUB_API = "https://api.charmhub.io/v2/charms/info"
CHARM_GITHUB_REPO = "canonical/k8s-operator"

# Launchpad snap owner/project path — builds live at ~containers/k8s/+snap/<name>
LP_SNAP_OWNER = "~containers"
LP_SNAP_PROJECT = "k8s"

METADATA_DIR = pathlib.Path(__file__).parent.parent / "metadata"

# PR number embedded in squash-merge commit subject, e.g. "fix: something (#2131)"
_PR_IN_SUBJECT_RE = re.compile(r"\(#(\d+)\)$")


# ---------------------------------------------------------------------------
# Snap pipeline helpers
# ---------------------------------------------------------------------------


def _snap_store_revision(track: str) -> int:
    """Return the current amd64 revision number published on *track*.

    The track argument should be the full channel name, e.g. '1.32-classic/stable'.
    The Snap Store channel-map uses '<track>/<risk>' in channel.name, where
    track contains the version + flavor (e.g. '1.32-classic').
    """
    resp = requests.get(
        SNAP_STORE_API,
        headers={"Snap-Device-Series": "16"},
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()

    # Normalise: accept both '1.32-classic/stable' and '1.32-classic' + risk split
    if "/" in track:
        track_name, risk = track.split("/", 1)
    else:
        raise ValueError(
            f"track must be in '<version>/<risk>' format, e.g. '1.32-classic/stable'. Got: {track!r}"
        )

    for entry in data.get("channel-map", []):
        ch = entry["channel"]
        if ch["track"] == track_name and ch["risk"] == risk and ch["architecture"] == "amd64":
            return int(entry["revision"])

    raise ValueError(
        f"Track '{track}' not found in Snap Store channel-map for 'k8s'. "
        f"Available stable tracks: {sorted({e['channel']['track'] for e in data.get('channel-map', []) if e['channel']['risk'] == 'stable'})}"
    )


def _launchpad_sha(track: str, revision: int) -> str:
    """Return the git SHA that produced *revision* for *track* via Launchpad builds.

    Launchpad snap builds live at:
      /devel/~containers/k8s/+snap/k8s-snap-<track>/builds

    Each build entry has:
      - store_upload_revision: the snap store revision number (int)
      - revision_id: the VCS commit SHA used for the build
    """
    # Derive the Launchpad snap name from the track, e.g. '1.32-classic/stable' -> 'k8s-snap-1.32-classic'
    track_name = track.split("/")[0]
    snap_name = f"k8s-snap-{track_name}"
    # URL pattern: /devel/~containers/k8s/+snap/<name>/builds
    builds_url = (
        f"{LAUNCHPAD_API}/{LP_SNAP_OWNER}/{LP_SNAP_PROJECT}/+snap/{snap_name}/builds"
    )

    # Paginate through builds (newest first) until we find the matching store revision
    url: str | None = f"{builds_url}?ws.size=75"
    while url:
        resp = requests.get(url, timeout=15)
        resp.raise_for_status()
        data = resp.json()
        for build in data.get("entries", []):
            if build.get("store_upload_revision") == revision:
                sha = build.get("revision_id")
                if sha:
                    return sha
                raise ValueError(
                    f"Launchpad build for revision {revision} found but 'revision_id' is empty. "
                    f"Build keys: {list(build.keys())}"
                )
        url = data.get("next_collection_link")

    raise ValueError(
        f"No Launchpad build found for snap '{snap_name}' with store revision {revision}. "
        "The build may still be in progress, or the snap name may have changed."
    )


# ---------------------------------------------------------------------------
# GitHub / shared helpers
# ---------------------------------------------------------------------------


def _github_commit_diff(sha: str, repo: str, headers: dict) -> str:
    """Return the unified diff for a single commit as a formatted string.

    Makes one API call to GET /repos/{repo}/commits/{sha}. The *headers*
    dict (auth token, Accept header, API version) is shared with the caller
    so it does not need to be rebuilt.
    """
    resp = requests.get(
        f"{GITHUB_API}/repos/{repo}/commits/{sha}",
        headers=headers,
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()
    return "\n\n".join(
        f"--- {f['filename']} ---\n{f.get('patch', '(binary or no diff)')}"
        for f in data.get("files", [])
    )


# GitHub caps per_page at 250 for an unpaginated compare request, but once a
# page/per_page parameter is supplied the endpoint's documented cap drops to 100.
_COMPARE_PAGE_SIZE = 100
# Safety cap on pages fetched (50k commits) so a misbehaving response can't loop forever.
_MAX_COMPARE_PAGES = 500


def _github_compare_commits(base_sha: str, head_sha: str, repo: str, headers: dict) -> list[dict[str, Any]]:
    """Return every commit between *base_sha* and *head_sha*, paginating as needed.

    Backfills or long gaps between runs (see README) can produce deltas larger
    than a single compare page, so pages are fetched until the aggregated
    commit count reaches total_commits.
    """
    commits: list[dict[str, Any]] = []
    page = 1
    data: dict[str, Any] = {}
    while True:
        resp = requests.get(
            f"{GITHUB_API}/repos/{repo}/compare/{base_sha}...{head_sha}"
            f"?per_page={_COMPARE_PAGE_SIZE}&page={page}",
            headers=headers,
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        page_commits = data.get("commits", [])
        commits.extend(page_commits)
        total_commits = data.get("total_commits")
        if not page_commits or (total_commits is not None and len(commits) >= total_commits):
            break
        page += 1
        if page > _MAX_COMPARE_PAGES:
            break

    total_commits = data.get("total_commits")
    if total_commits is not None and total_commits > len(commits):
        raise ValueError(
            f"GitHub compare returned {len(commits)} of {total_commits} commits "
            f"for {repo} {base_sha}...{head_sha} after paginating. Narrow the delta."
        )
    return commits


def _github_commits(base_sha: str, head_sha: str, repo: str = GITHUB_REPO) -> list[dict[str, Any]]:
    """Return one entry per commit between *base_sha* and *head_sha*.

    Uses paginated GitHub compare requests to obtain the commit list, then
    fetches each commit's individual diff via a dedicated API call so the AI
    receives focused, accurate context instead of the entire release's
    aggregate diff. PR numbers are extracted from commit message subjects
    where GitHub embeds them (e.g. "fix: something (#2131)").
    """
    token = os.environ.get("GITHUB_TOKEN")
    headers = {"Authorization": f"Bearer {token}"} if token else {}
    headers["Accept"] = "application/vnd.github+json"
    headers["X-GitHub-Api-Version"] = "2022-11-28"

    commits = _github_compare_commits(base_sha, head_sha, repo, headers)

    if not token and len(commits) > 20:
        print(
            f"\u26a0 GITHUB_TOKEN not set \u2014 fetching per-commit diffs for {len(commits)} commits "
            "against the 60 req/hr unauthenticated rate limit. "
            "Set GITHUB_TOKEN to avoid throttling.",
            file=sys.stderr,
        )

    entries: list[dict[str, Any]] = []
    for commit in commits:
        msg_lines = commit["commit"]["message"].splitlines()
        title = msg_lines[0]
        body = "\n".join(msg_lines[1:]).strip()

        m = _PR_IN_SUBJECT_RE.search(title)
        pr_number = int(m.group(1)) if m else None
        pr_url = (
            f"https://github.com/{repo}/pull/{pr_number}"
            if pr_number else None
        )

        diff = _github_commit_diff(commit["sha"], repo, headers)

        entries.append({
            "sha": commit["sha"],
            "title": title,
            "body": body,
            "html_url": commit.get("html_url", ""),
            "author": (commit.get("author") or {}).get("login")
                      or commit["commit"]["author"]["name"],
            "date": commit["commit"]["author"]["date"][:10],
            "pr_number": pr_number,
            "pr_url": pr_url,
            "diff": diff,
        })

    return entries


# ---------------------------------------------------------------------------
# Charm pipeline helpers
# ---------------------------------------------------------------------------


def _charmhub_revision(track: str) -> int:
    """Return the current amd64 revision number published on *track* for the k8s charm.

    The track argument should be the full channel name, e.g. '1.32/stable'.
    """
    resp = requests.get(
        f"{CHARMHUB_API}/k8s?fields=channel-map",
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()

    for entry in data.get("channel-map", []):
        ch = entry["channel"]
        if (
            ch.get("name") == track
            and entry["revision"]["bases"][0].get("architecture") == "amd64"
        ):
            return int(entry["revision"]["revision"])

    available = sorted(
        {e["channel"]["name"] for e in data.get("channel-map", []) if e["channel"].get("risk") == "stable"}
    )
    raise ValueError(
        f"Track '{track}' not found in Charmhub channel-map for 'k8s'. "
        f"Available stable tracks: {available}"
    )


def _charm_github_sha(revision: int) -> str:
    """Return the git SHA for *revision* of the k8s charm via its GitHub tag.

    Tags follow the pattern 'k8s-rev{revision}' in canonical/k8s-operator.
    Tags are lightweight (type 'commit'), so object.sha is the commit SHA directly.
    """
    token = os.environ.get("GITHUB_TOKEN")
    headers = {"Authorization": f"Bearer {token}"} if token else {}
    headers["Accept"] = "application/vnd.github+json"
    headers["X-GitHub-Api-Version"] = "2022-11-28"

    tag_name = f"k8s-rev{revision}"
    resp = requests.get(
        f"{GITHUB_API}/repos/{CHARM_GITHUB_REPO}/git/refs/tags/{tag_name}",
        headers=headers,
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()
    sha = data["object"]["sha"]
    return sha


# ---------------------------------------------------------------------------
# Storage helpers
# ---------------------------------------------------------------------------


def delta_path(channel_key: str) -> pathlib.Path:
    """Return the path for a track's saved delta file."""
    safe = channel_key.replace(":", "-").replace("/", "-")
    return METADATA_DIR / f"delta-{safe}.json"


def _save_delta(channel_key: str, prs: list[dict[str, Any]]) -> None:
    """Persist delta to metadata/delta-<safe-track>.json."""
    METADATA_DIR.mkdir(exist_ok=True)
    path = delta_path(channel_key)
    path.write_text(json.dumps(prs, indent=2))


def load_delta(channel_key: str) -> list[dict[str, Any]]:
    """Load a previously saved delta from disk."""
    path = delta_path(channel_key)
    if not path.exists():
        raise FileNotFoundError(
            f"No delta file found for '{channel_key}'. Run `fetch` first."
        )
    return json.loads(path.read_text())


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def fetch_charm_delta(channel_key: str) -> list[dict[str, Any]]:
    """Full charm pipeline: Charmhub -> GitHub tag -> GitHub commits. Returns commit list.

    *channel_key* must be in the form 'charm:<channel>', e.g. 'charm:1.32/stable'.
    Only '/stable' channels are accepted.
    """
    if not channel_key.startswith("charm:"):
        raise ValueError(
            f"channel_key must start with 'charm:', got: {channel_key!r}"
        )
    channel = channel_key[len("charm:"):]
    if not channel.endswith("/stable"):
        raise ValueError(
            f"Charm patch notices only support '/stable' channels, got: {channel!r}"
        )

    channel_metadata = metadata.load().get("tracks", {}).get(channel_key, {})
    base_sha = channel_metadata.get("last_documented_sha")
    if not base_sha:
        raise ValueError(
            f"No last_documented_sha for '{channel_key}' in patch-metadata.json. "
            "Add an initial entry before running fetch."
        )

    revision = _charmhub_revision(channel)
    head_sha = _charm_github_sha(revision)
    commits = _github_commits(base_sha, head_sha, repo=CHARM_GITHUB_REPO)
    _save_delta(channel_key, commits)
    return commits


def fetch_snap_delta(channel_key: str) -> list[dict[str, Any]]:
    """Full pipeline: Snap Store -> Launchpad -> GitHub. Returns PR list."""
    if not channel_key.startswith("snap:"):
        raise ValueError(
            f"channel_key must start with 'snap:', got: {channel_key!r}"
        )
    track = channel_key[len("snap:"):]
    channel_metadata = metadata.load().get("tracks", {}).get(channel_key, {})
    base_sha = channel_metadata.get("last_documented_sha")
    if not base_sha:
        raise ValueError(
            f"No last_documented_sha for '{channel_key}' in patch-metadata.json. "
            "Add an initial entry before running fetch."
        )
    revision = _snap_store_revision(track)
    head_sha = _launchpad_sha(track, revision)
    prs = _github_commits(base_sha, head_sha)
    _save_delta(channel_key, prs)
    return prs