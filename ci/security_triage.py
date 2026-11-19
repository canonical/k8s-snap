#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Security PoC daily triage script.

Fetches code scanning alerts from Canonical Kubernetes repos,
tracks new vs known alerts, and posts a structured triage summary
to Mattermost via incoming webhook.

The goal is to replace manual GitHub security page triage — this report
is the single source of truth for open alerts.

Usage:
    python3 ci/security_triage.py --webhook <URL> [--dry-run]

Environment variables:
    GH_TOKEN  — GitHub token with security_events scope for all target repos
"""

import argparse
import json
import os
import re
import subprocess
import sys
from collections.abc import Iterable
from datetime import datetime, timezone
from pathlib import Path
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

PRIMARY_REPO = "canonical/k8s-snap"
REPOS = [
    "canonical/k8s-snap",
    "canonical/k8sd",
    "canonical/k8s-snap-api",
]

# Alerts older than this are flagged as overdue
OVERDUE_DAYS = 7

STATE_VERSION = 2

# State file for delta detection (stored as GitHub Actions cache or local file)
STATE_FILE = Path(os.environ.get("STATE_FILE", "/tmp/security-triage-state.json"))


class GhApiError(Exception):

    def __init__(self, endpoint: str, stderr: str, returncode: int = 1):
        super().__init__(f"failed to fetch {endpoint}: {stderr}")
        self.endpoint = endpoint
        self.stderr = stderr
        self.returncode = returncode

    @property
    def is_permission_or_not_found(self) -> bool:
        return "403" in self.stderr or "404" in self.stderr


def gh_api(endpoint: str) -> list[dict[str, Any]]:
    """Call gh api and return parsed JSON."""
    cmd = ["gh", "api", endpoint, "--paginate", "--slurp"]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        pages = json.loads(result.stdout)
        return [
            item
            for page in pages
            for item in (page if isinstance(page, list) else [page])
        ]
    except subprocess.CalledProcessError as e:
        raise GhApiError(endpoint, e.stderr, e.returncode) from e
    except json.JSONDecodeError as e:
        raise GhApiError(endpoint, f"JSON decode error: {e}") from e


def days_ago(iso_date: str) -> int:
    """Return number of days since the given ISO date string."""
    created = datetime.fromisoformat(iso_date.replace("Z", "+00:00"))
    return (datetime.now(timezone.utc) - created).days


def load_previous_state() -> dict[str, Any]:
    """Load state from the previous run."""
    if STATE_FILE.exists():
        try:
            data = json.loads(STATE_FILE.read_text())
            if data.get("version", 1) < STATE_VERSION:
                print(
                    "State file is from an older schema version; resetting state.",
                    file=sys.stderr,
                )
                return {}
            return data
        except (json.JSONDecodeError, KeyError):
            pass
    return {}


def save_state(alert_ids: set[str]) -> None:
    """Persist current alert IDs for next run's delta comparison."""
    STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
    STATE_FILE.write_text(
        json.dumps(
            {
                "version": STATE_VERSION,
                "alert_ids": sorted(alert_ids),
                "updated": datetime.now(timezone.utc).isoformat(),
            },
        )
    )


def _normalize_severity(rule: dict[str, Any]) -> str:
    """Map code scanning alert severity to a uniform level string."""
    sev = rule.get("security_severity_level", "") or ""
    if sev:
        return sev.lower()
    raw = rule.get("severity", "none").lower()
    if raw == "note":
        return "note"
    return {"error": "high", "warning": "medium", "low": "low"}.get(raw, "none")


def _description_from_rule(rule: dict[str, Any]) -> str:
    """Extract a short description from a code scanning rule."""
    desc = rule.get("description") or rule.get("full_description") or ""
    if desc:
        return desc
    help_text = rule.get("help", "")
    if not help_text:
        return ""
    first_line = help_text.split("\n", 1)[0].strip()
    if first_line.startswith("`"):
        first_line = first_line.split("`:", 1)[-1].strip().lstrip("`")
    return first_line


def fetch_release_branches(repo: str) -> list[str]:
    """Fetch active release-* branches for a repo."""
    data = gh_api(f"/repos/{repo}/branches?per_page=100")
    pattern = re.compile(r"^release-[0-9]+\.[0-9]+$")
    branches = [
        item["name"]
        for item in data
        if isinstance(item, dict) and "name" in item and pattern.match(item["name"])
    ]

    def parse_version(b: str) -> tuple[int, ...]:
        try:
            return tuple(int(x) for x in b.removeprefix("release-").split("."))
        except ValueError:
            return (0,)

    return sorted(branches, key=parse_version)


def compact_branches(branches: Iterable[str]) -> str:
    """Format and compact a collection of branch names.

    e.g. ['main', 'release-1.32', 'release-1.33', 'release-1.34', 'release-1.35']
    -> 'main, release-1.32..1.35'
    """
    branch_set = set(branches)
    if not branch_set:
        return "—"

    has_main = "main" in branch_set
    non_release = sorted(
        b for b in branch_set if b != "main" and not b.startswith("release-")
    )

    # Parse release branches: (major, minor)
    rel_branches: list[tuple[int, int]] = []
    for b in branch_set:
        if b.startswith("release-"):
            parts = b.removeprefix("release-").split(".")
            if len(parts) == 2 and all(p.isdigit() for p in parts):
                rel_branches.append((int(parts[0]), int(parts[1])))

    rel_branches.sort()

    ranges: list[str] = []
    if rel_branches:
        cur_major, cur_start = rel_branches[0]
        cur_end = cur_start
        for major, minor in rel_branches[1:]:
            if major == cur_major and minor == cur_end + 1:
                cur_end = minor
            else:
                if cur_start == cur_end:
                    ranges.append(f"release-{cur_major}.{cur_start}")
                else:
                    ranges.append(
                        f"release-{cur_major}.{cur_start}..{cur_major}.{cur_end}"
                    )
                cur_major, cur_start, cur_end = major, minor, minor
        if cur_start == cur_end:
            ranges.append(f"release-{cur_major}.{cur_start}")
        else:
            ranges.append(f"release-{cur_major}.{cur_start}..{cur_major}.{cur_end}")

    result: list[str] = []
    if has_main:
        result.append("main")
    result.extend(non_release)
    result.extend(ranges)
    return ", ".join(result) if result else "—"


def fetch_code_scanning_alerts(repo: str) -> list[dict[str, Any]]:
    """Fetch open code scanning alerts for a repo across default and release branches."""
    branches = fetch_release_branches(repo)

    # 1. Fetch default branch query
    raw_default = gh_api(f"/repos/{repo}/code-scanning/alerts?state=open&per_page=100")

    # 2. Fetch release branches
    branch_queries: list[tuple[str, list[dict[str, Any]]]] = []
    for b in branches:
        raw_b = gh_api(
            f"/repos/{repo}/code-scanning/alerts?state=open&ref=refs/heads/{b}&per_page=100"
        )
        branch_queries.append((b, raw_b))

    alerts_by_number: dict[int, dict[str, Any]] = {}

    def record_alert(raw_alert: dict[str, Any], branch_name: str) -> None:
        alert_number = raw_alert.get("number")
        if not alert_number:
            return
        if alert_number not in alerts_by_number:
            rule = raw_alert.get("rule", {})
            severity = _normalize_severity(rule)
            rule_id = rule.get("id", "unknown")
            tool = raw_alert.get("tool", {}).get("name", "unknown")
            created = raw_alert.get("created_at", "")
            age = days_ago(created) if created else 0
            description = _description_from_rule(rule)
            alert_link = (
                f"https://github.com/{repo}/security/code-scanning/{alert_number}"
            )

            alerts_by_number[alert_number] = {
                "id": f"{repo}#{alert_number}",
                "alert_number": alert_number,
                "rule_id": rule_id,
                "severity": severity,
                "tool": tool,
                "age_days": age,
                "summary": description,
                "link": alert_link,
                "repo": repo,
                "branches": set(),
            }
        alerts_by_number[alert_number]["branches"].add(branch_name)

    for a in raw_default:
        ref = a.get("most_recent_instance", {}).get("ref", "")
        branch = (
            ref.removeprefix("refs/heads/") if ref.startswith("refs/heads/") else "main"
        )
        record_alert(a, branch or "main")

    for b, raw_b in branch_queries:
        for a in raw_b:
            record_alert(a, b)

    alerts: list[dict[str, Any]] = []
    for a in alerts_by_number.values():
        a["branches_display"] = compact_branches(a["branches"])
        alerts.append(a)

    return alerts


def severity_sort_key(alert: dict[str, Any]) -> tuple[int, int]:
    """Sort key: highest severity first, then oldest first."""
    rank = {"critical": 5, "high": 4, "medium": 3, "low": 2, "note": 1}.get(
        alert["severity"].lower(), 0
    )
    return (-rank, -alert["age_days"])


def sev_emoji(severity: str) -> str:
    return {
        "critical": "🔥",
        "high": "🔴",
        "medium": "🟡",
        "low": "🟢",
        "note": "ℹ️",
    }.get(severity.lower(), "⚪")


def format_alert_row(a: dict[str, Any]) -> str:
    """Format one alert as a markdown table row."""
    sev = a["severity"]
    sev_display = f"{sev_emoji(sev)} {sev.capitalize()}" if sev else "⚪ —"
    summary = a.get("summary", "").replace("\n", " ").replace("|", "\\|")
    if len(summary) > 55:
        summary = summary[:52] + "..."
    link_display = f"[{a['rule_id']}]({a['link']})"
    overdue = (
        " ⏰"
        if a["age_days"] > OVERDUE_DAYS
        and a["severity"].lower() in ("critical", "high")
        else ""
    )
    repo_display = a.get("repo", "").split("/")[-1]
    branch_display = a.get("branches_display", "—")
    return (
        f"| {repo_display} | {link_display} | {a['tool']} | {sev_display} | "
        f"{branch_display} | {summary} | {a['age_days']}d{overdue} |"
    )


def build_message(
    all_alerts: list[dict[str, Any]],
    new_ids: set[str],
    resolved_ids: set[str],
    is_first_run: bool,
) -> dict[str, Any]:
    """Build the full Mattermost webhook payload with new/known/resolved sections."""
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")

    actionable_alerts = [a for a in all_alerts if a["severity"].lower() != "note"]
    note_alerts = [a for a in all_alerts if a["severity"].lower() == "note"]

    new_alerts = [a for a in actionable_alerts if a["id"] in new_ids]
    known_alerts = [a for a in actionable_alerts if a["id"] not in new_ids]

    new_alerts.sort(key=severity_sort_key)
    known_alerts.sort(key=severity_sort_key)

    # Header stats
    total = len(actionable_alerts)
    n_new = len(new_alerts)
    n_resolved = len(resolved_ids)
    n_critical = sum(
        1 for a in actionable_alerts if a["severity"].lower() == "critical"
    )
    n_high = sum(1 for a in actionable_alerts if a["severity"].lower() == "high")
    n_notes = len(note_alerts)

    lines = []
    lines.append(
        f"📊 **{total}** open"
        + (f" · 🆕 **{n_new}** new" if n_new else "")
        + (f" · ✅ **{n_resolved}** resolved" if n_resolved else "")
        + (f" · 🔥 {n_critical} critical" if n_critical else "")
        + (f" · 🔴 {n_high} high" if n_high else "")
        + (f" · ℹ️ {n_notes} notes" if n_notes else "")
    )
    lines.append("")

    # NEW section — always expanded, full detail
    if new_alerts:
        lines.append("### 🆕 NEW — needs attention")
        lines.append("| Repo | Rule | Tool | Severity | Branch | Description | Age |")
        lines.append("|------|------|------|----------|--------|-------------|-----|")
        for a in new_alerts:
            lines.append(format_alert_row(a))
        lines.append("")
    elif not is_first_run:
        lines.append("### 🆕 No new alerts since last run")
        lines.append("")

    # KNOWN section — expandable by tool
    if known_alerts:
        # Group by tool
        by_tool: dict[str, list[dict[str, Any]]] = {}
        for a in known_alerts:
            by_tool.setdefault(a["tool"], []).append(a)

        lines.append("### 📋 KNOWN — previously reported")
        for tool_name, tool_alerts in sorted(by_tool.items()):
            tool_alerts.sort(key=severity_sort_key)
            sev_summary = _severity_summary(tool_alerts)
            lines.append(
                f"#### {tool_name} ({len(tool_alerts)} alerts) — {sev_summary}"
            )
            lines.append(
                "| Repo | Rule | Tool | Severity | Branch | Description | Age |"
            )
            lines.append(
                "|------|------|------|----------|--------|-------------|-----|"
            )
            for a in tool_alerts:
                lines.append(format_alert_row(a))
            lines.append("")

    # RESOLVED section
    if resolved_ids:
        lines.append("### ✅ RESOLVED — gone since last run")
        for rid in sorted(resolved_ids):
            lines.append(f"- ~{rid}~")
        lines.append("")

    # Informational notes summary
    if note_alerts:
        lines.append(f"ℹ️ _{n_notes} informational note(s) omitted from tables._\n")

    # Footer
    repos_str = ", ".join(f"`{r}`" for r in REPOS)
    workflow_url = (
        "https://github.com/canonical/k8s-snap/actions/workflows/security-triage.yaml"
    )
    lines.append(f"---\n_Repos: {repos_str} · [Workflow run]({workflow_url})_")

    text = "\n".join(lines)

    # Color based on severity
    if n_critical > 0 or any(a["severity"].lower() == "critical" for a in new_alerts):
        color = "#FF0000"
        status_emoji = "🚨"
    elif n_high > 0 or new_alerts:
        color = "#FFA500"
        status_emoji = "⚠️"
    else:
        color = "#36A64F"
        status_emoji = "✅"

    title = f"{status_emoji} Security Triage — {today}"

    return {
        "attachments": [
            {
                "fallback": f"Security Triage {today}: {total} alerts, {n_new} new, {n_resolved} resolved",
                "title": title,
                "text": text,
                "color": color,
            }
        ]
    }


def _severity_summary(alerts: list[dict[str, Any]]) -> str:
    """One-line severity breakdown for a group."""
    counts: dict[str, int] = {}
    for a in alerts:
        sev = a["severity"].lower() if a["severity"] else "unrated"
        counts[sev] = counts.get(sev, 0) + 1
    parts = []
    for sev in ("critical", "high", "medium", "low", "note", "unrated"):
        if sev in counts:
            parts.append(
                f"{sev_emoji(sev) if sev != 'unrated' else '⚪'} {counts[sev]} {sev}"
            )
    return ", ".join(parts) if parts else "—"


def post_webhook(webhook_url: str, payload: dict[str, Any]) -> None:
    """Post payload to Mattermost incoming webhook."""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = Request(
        webhook_url,
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8"},
    )
    try:
        with urlopen(req, timeout=30) as resp:
            resp.read()
        print("Posted to Mattermost successfully.", file=sys.stderr)
    except (HTTPError, URLError) as e:
        print(f"Webhook error: {e}", file=sys.stderr)
        raise SystemExit(2)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Fetch security alerts and post triage summary to Mattermost"
    )
    parser.add_argument(
        "--webhook",
        "-w",
        default=None,
        help="Mattermost incoming webhook URL (or set MATTERMOST_BOT_WEBHOOK_URL)",
    )
    parser.add_argument(
        "--channel-id",
        default=None,
        help="Channel ID to post to (or set MATTERMOST_CHANNEL_ID)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print the payload without posting",
    )
    parser.add_argument(
        "--full",
        action="store_true",
        help="Always post (even if no delta)",
    )
    args = parser.parse_args()

    webhook = args.webhook or os.environ.get("MATTERMOST_BOT_WEBHOOK_URL")
    channel_id = args.channel_id or os.environ.get("MATTERMOST_CHANNEL_ID")

    if not webhook and not args.dry_run:
        print(
            "Error: --webhook or MATTERMOST_BOT_WEBHOOK_URL required", file=sys.stderr
        )
        return 1

    # Fetch all alerts from all repos
    all_alerts: list[dict[str, Any]] = []
    successful_repos: set[str] = set()

    for repo in REPOS:
        print(f"Fetching alerts for {repo}...", file=sys.stderr)
        try:
            alerts = fetch_code_scanning_alerts(repo)
            all_alerts.extend(alerts)
            successful_repos.add(repo)
            print(f"  {repo}: {len(alerts)} open alerts", file=sys.stderr)
        except GhApiError as e:
            if repo != PRIMARY_REPO and e.is_permission_or_not_found:
                print(
                    f"Warning: skipping {repo}: insufficient permissions or repo not found ({e.stderr.strip()})",
                    file=sys.stderr,
                )
                continue
            print(
                f"Error: failed to fetch alerts for {repo}: {e.stderr.strip()}",
                file=sys.stderr,
            )
            raise SystemExit(1)

    if not successful_repos or PRIMARY_REPO not in successful_repos:
        print(
            "Error: primary repo failed or all repos failed to fetch",
            file=sys.stderr,
        )
        return 1

    # Delta detection
    actionable_alerts = [a for a in all_alerts if a["severity"].lower() != "note"]
    note_alerts = [a for a in all_alerts if a["severity"].lower() == "note"]
    note_ids = {a["id"] for a in note_alerts}

    current_ids = {a["id"] for a in actionable_alerts}
    prev_state = load_previous_state()
    previous_ids = set(prev_state.get("alert_ids", []))
    is_first_run = not prev_state

    skipped_repos = set(REPOS) - successful_repos
    skipped_previous_ids = {
        rid
        for rid in previous_ids
        if any(rid.startswith(f"{repo}#") for repo in skipped_repos)
    }
    evaluated_previous_ids = previous_ids - skipped_previous_ids

    new_ids = current_ids - evaluated_previous_ids
    resolved_ids = (evaluated_previous_ids - current_ids) - note_ids

    has_delta = bool(new_ids or resolved_ids)
    if not has_delta and not args.full and not is_first_run:
        if args.dry_run:
            print(
                json.dumps(
                    {
                        "skipped": True,
                        "reason": "no delta",
                        "total_open": len(actionable_alerts),
                    }
                )
            )
        return 0

    # Build message
    payload = build_message(all_alerts, new_ids, resolved_ids, is_first_run)

    if channel_id:
        payload["channel_id"] = channel_id

    if args.dry_run:
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        return 0

    post_webhook(webhook, payload)
    # Save state only after a successful post so a delivery failure does not
    # suppress the delta notification on the next run. Preserve IDs of skipped repos.
    save_state(current_ids | skipped_previous_ids)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
