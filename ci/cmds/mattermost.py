#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Mattermost subcommands for `k8s-ci`.

This module provides functionality to post aggregated CI results to a Mattermost
channel via incoming webhooks or bot accounts. It supports
posting a concise summary via webhook (with color support) and then posting
detailed results as a threaded comment using a bot.
"""

import argparse
import json
import os
import sys
from typing import Any, Dict, List, Optional, Tuple
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

MAX_FAILING_TESTS = 5
_NON_FAILING_STATUSES = {"success", "skipped"}


def add_mattermost_cmds(parser: argparse.ArgumentParser) -> None:
    """
    Register Mattermost-related subcommands to the given CLI parser.

    Args:
        parser: The parent argparse.ArgumentParser to which subcommands will be added.
    """
    mattermost_parser = parser.add_parser(
        "mattermost", help="Post aggregated CI results to Mattermost."
    )
    mattermost_sub = mattermost_parser.add_subparsers(
        dest="mattermost_command", required=True, title="mattermost commands"
    )

    # Results-message command
    p = mattermost_sub.add_parser(
        "results-message",
        help="Aggregate a list of results and post to Mattermost.",
    )
    p.add_argument(
        "--file", "-f", default=None, help="Results file (JSON) or '-' for stdin."
    )
    p.add_argument(
        "--webhook", "-w", help="Incoming webhook URL (or MATTERMOST_WEBHOOK_URL)"
    )
    p.add_argument("--title", "-t", default=None, help="Message title")
    p.add_argument("--dry-run", action="store_true", help="Print payload and exit")
    p.add_argument(
        "--bot-token", default=None, help="Bot token (or MATTERMOST_BOT_TOKEN)"
    )
    p.add_argument("--server", default=None, help="Server URL (or MATTERMOST_SERVER)")
    p.add_argument(
        "--channel-id", default=None, help="Channel ID (or MATTERMOST_CHANNEL_ID)"
    )
    p.set_defaults(func=cmd_results_message)


def _load_flattened_json(path: Optional[str]) -> List[Dict[str, Any]]:
    """
    Load JSON results from a file or stdin and flatten nested arrays.

    Args:
        path: File path or '-' for stdin.

    Returns:
        List of flattened result entries.

    Raises:
        SystemExit if input is invalid or missing.
    """
    if path == "-":
        data = json.load(sys.stdin)
    elif path:
        with open(path, "r", encoding="utf-8") as fh:
            data = json.load(fh)
    else:
        raise SystemExit("Error: must provide --file or '-' for stdin")

    # Flatten nested lists
    if isinstance(data, list) and len(data) == 1 and isinstance(data[0], list):
        data = data[0]

    if isinstance(data, list) and any(isinstance(x, list) for x in data):
        flat: List[Any] = []
        for x in data:
            flat.extend(x if isinstance(x, list) else [x])
        data = flat

    if not isinstance(data, list):
        raise SystemExit(f"{path or 'stdin'} must be an array of objects")

    return data


def _determine_run_link(entry: Dict[str, Any]) -> Optional[str]:
    """Return the URL for the CI run associated with an entry, if available."""
    run_url = entry.get("run_url") or entry.get("runLink") or entry.get("run_link")
    if run_url:
        return str(run_url)
    repo = os.environ.get("GITHUB_REPOSITORY")
    run_id = os.environ.get("GITHUB_RUN_ID")
    if repo and run_id:
        return f"https://github.com/{repo}/actions/runs/{run_id}"
    return None


def _entry_channel(entry: Dict[str, Any]) -> str:
    """Return the channel label for a result entry."""
    channel = str(entry.get("channel") or "").strip()
    if channel:
        return channel
    artifact = str(entry.get("artifact") or "").strip()
    if artifact:
        return f"artifact:{artifact}"
    return "unknown"


def _entry_status(entry: Dict[str, Any]) -> str:
    """Return the normalized lowercase status for an entry."""
    return str(entry.get("status") or "").strip().lower()


def _format_os_arch_header(os_name: str, arch: str) -> str:
    """Format an OS and architecture pair as a compact table column header."""
    short_os = os_name.removeprefix("ubuntu:")
    return f"{short_os} {arch}"


def _format_cell(cell_entries: List[Dict[str, Any]]) -> str:
    """Format a single channel x (os, arch) matrix cell."""
    if not cell_entries:
        return "—"
    non_passing = [
        e for e in cell_entries if _entry_status(e) not in _NON_FAILING_STATUSES
    ]
    if not non_passing:
        return ":white_check_mark:"
    cancelled = sum(1 for e in non_passing if _entry_status(e) == "cancelled")
    failed = len(non_passing) - cancelled
    parts: List[str] = []
    if failed:
        parts.append(f":x: {failed}")
    if cancelled:
        parts.append(f":stopwatch: {cancelled}")
    return " ".join(parts)


def _format_status_breakdown(failed_entries: List[Dict[str, Any]]) -> str:
    """Format status counts (failed vs cancelled) for a failing test row."""
    cancelled = sum(1 for e in failed_entries if _entry_status(e) == "cancelled")
    failed = len(failed_entries) - cancelled
    parts: List[str] = []
    if failed:
        parts.append(f":x: {failed}")
    if cancelled:
        parts.append(f":stopwatch: {cancelled}")
    return ", ".join(parts)


def _build_channel_matrix_table(
    entries: List[Dict[str, Any]], channels: List[str]
) -> List[str]:
    """Build the Channel Matrix markdown table lines."""
    targets: List[Tuple[str, str]] = sorted(
        {
            (str(e.get("os") or "unknown"), str(e.get("arch") or "unknown"))
            for e in entries
        }
    )
    grid: Dict[Tuple[str, str, str], List[Dict[str, Any]]] = {}
    for e in entries:
        key = (
            _entry_channel(e),
            str(e.get("os") or "unknown"),
            str(e.get("arch") or "unknown"),
        )
        grid.setdefault(key, []).append(e)

    headers = [_format_os_arch_header(os_name, arch) for os_name, arch in targets]
    lines = [
        "### Channel Matrix",
        "| Channel | " + " | ".join(headers) + " |",
        "| --- | " + " | ".join(":---:" for _ in targets) + " |",
    ]
    for ch in channels:
        cells = [
            _format_cell(grid.get((ch, os_name, arch), [])) for os_name, arch in targets
        ]
        lines.append(f"| `{ch}` | " + " | ".join(cells) + " |")
    return lines


def _build_failing_tests_table(
    entries: List[Dict[str, Any]], all_channels: List[str]
) -> List[str]:
    """Build the Top Failing Tests markdown table lines, capped at MAX_FAILING_TESTS."""
    by_node_id: Dict[str, List[Dict[str, Any]]] = {}
    for e in entries:
        if _entry_status(e) in _NON_FAILING_STATUSES:
            continue
        node_id = str(e.get("test") or "").strip() or "Unnamed Test"
        by_node_id.setdefault(node_id, []).append(e)

    if not by_node_id:
        return []

    short_counts: Dict[str, int] = {}
    for node_id in by_node_id:
        short = node_id.split("::")[-1]
        short_counts[short] = short_counts.get(short, 0) + 1

    def _display_name(node_id: str) -> str:
        short = node_id.split("::")[-1]
        if short_counts.get(short, 0) > 1 and "::" in node_id:
            return node_id.removeprefix("tests/")
        return short

    ranked = sorted(
        by_node_id.items(), key=lambda kv: (-len(kv[1]), _display_name(kv[0]), kv[0])
    )
    channel_order = {ch: idx for idx, ch in enumerate(all_channels)}

    lines = [
        "### Top Failing Tests",
        "| Test | Count | Breakdown | Channels |",
        "| --- | ---: | --- | --- |",
    ]
    for node_id, test_entries in ranked[:MAX_FAILING_TESTS]:
        safe_name = _display_name(node_id).replace("|", "\\|")
        affected = sorted(
            {_entry_channel(e) for e in test_entries},
            key=lambda ch: channel_order.get(ch, len(channel_order)),
        )
        if len(all_channels) >= 4 and len(affected) == len(all_channels):
            channels_display = f"all {len(all_channels)} channels"
        else:
            channels_display = ", ".join(f"`{ch}`" for ch in affected)
        breakdown = _format_status_breakdown(test_entries)
        lines.append(
            f"| `{safe_name}` | {len(test_entries)} | {breakdown} | {channels_display} |"
        )

    remaining = len(ranked) - MAX_FAILING_TESTS
    if remaining > 0:
        noun = "test" if remaining == 1 else "tests"
        more_label = f"{remaining} more failing {noun}"
        run_link = next(
            (
                _determine_run_link(e)
                for _, es in ranked[MAX_FAILING_TESTS:]
                for e in es
                if _determine_run_link(e)
            ),
            None,
        )
        if run_link:
            lines.append(f"_...and [{more_label}]({run_link})._")
        else:
            lines.append(f"_...and {more_label}._")
    return lines


def build_details_message(entries: List[Dict[str, Any]]) -> str:
    """Build channel matrix and top failing test tables for a Mattermost thread reply."""
    if not entries:
        return "_No test results recorded._"

    channels = sorted({_entry_channel(e) for e in entries}, reverse=True)
    lines = _build_channel_matrix_table(entries, channels)
    failing_lines = _build_failing_tests_table(entries, channels)
    if failing_lines:
        lines.append("")
        lines.extend(failing_lines)
    return "\n".join(lines)


def _build_summary_payload(
    channel_id: str, entries: List[Dict[str, Any]], title: str, color: str
) -> Dict[str, Any]:
    """
    Build a concise summary payload for an incoming webhook.

    Args:
        channel_id: Channel ID to post to.
        entries: List of result entries.
        title: Title of the message.
        color: Color code ('good', 'warning', 'danger').

    Returns:
        Payload dictionary ready to post.
    """
    total = len(entries)
    successes = sum(1 for e in entries if _entry_status(e) == "success")
    skipped = sum(1 for e in entries if _entry_status(e) == "skipped")
    cancelled = sum(1 for e in entries if _entry_status(e) == "cancelled")
    failed = total - successes - skipped - cancelled
    non_passing = failed + cancelled
    pass_pct = (successes / total * 100) if total else 0.0
    emoji = ":white_check_mark:" if non_passing == 0 else ":x:"
    title = f"{emoji} {title}"

    parts = [f"{successes}/{total} passed ({pass_pct:.0f}%)"]
    if failed:
        parts.append(f":x: {failed} failed")
    if cancelled:
        parts.append(f":stopwatch: {cancelled} cancelled")
    parts.append(f"Skipped: {skipped}")
    text = " · ".join(parts)

    run_link = _determine_run_link(entries[0]) if entries else None
    if run_link:
        text = f"{text} — [View Run]({run_link})"

    return {
        "channel_id": channel_id,
        "attachments": [
            {"fallback": text, "title": title, "text": text, "color": color}
        ],
    }


def _determine_color(entries: List[Dict[str, Any]]) -> str:
    """Return 'good' when all entries passed or skipped, else 'danger'."""
    for e in entries:
        if _entry_status(e) not in _NON_FAILING_STATUSES:
            return "danger"
    return "good"


def post_webhook(webhook: str, payload: Dict[str, Any]) -> None:
    """Post payload to Mattermost using an incoming webhook."""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = Request(
        webhook, data=body, headers={"Content-Type": "application/json; charset=utf-8"}
    )
    try:
        with urlopen(req, timeout=20) as resp:
            resp.read()  # webhook returns plain "ok"
    except HTTPError as e:
        detail = e.read().decode("utf-8", errors="replace") if e.fp else ""
        print(f"Webhook error: {e} {detail}".strip(), file=sys.stderr)
        raise SystemExit(2)
    except URLError as e:
        print(f"Webhook error: {e}", file=sys.stderr)
        raise SystemExit(2)


def _post_bot(
    server: str,
    token: str,
    channel_id: str,
    message: str,
    root_id: Optional[str] = None,
) -> dict:
    """
    Post a message using a Mattermost bot account.

    Args:
        server: Base server URL.
        token: Bot token.
        channel_id: Channel to post in.
        message: Message content.
        root_id: Optional parent post ID to thread under.

    Returns:
        JSON response from server.
    """
    body = {"channel_id": channel_id, "message": message}
    if root_id:
        body["root_id"] = root_id
    req = Request(
        f"{server.rstrip('/')}/api/v4/posts",
        data=json.dumps(body).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
    )
    try:
        with urlopen(req, timeout=20) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except HTTPError as e:
        detail = e.read().decode("utf-8", errors="replace") if e.fp else ""
        print(f"Bot post error: {e} {detail}".strip(), file=sys.stderr)
        raise SystemExit(2)
    except URLError as e:
        print(f"Bot post error: {e}", file=sys.stderr)
        raise SystemExit(2)


def cmd_results_message(args: argparse.Namespace) -> int:
    """Command to post CI results summary via webhook and details as bot-threaded comment."""
    entries = _load_flattened_json(args.file)
    details_text = build_details_message(entries)
    color = _determine_color(entries)

    webhook = args.webhook or os.environ.get("MATTERMOST_WEBHOOK_URL")
    token = args.bot_token or os.environ.get("MATTERMOST_BOT_TOKEN")
    server = args.server or os.environ.get("MATTERMOST_SERVER")
    channel_id = args.channel_id or os.environ.get("MATTERMOST_CHANNEL_ID")
    title = args.title or "CI results"

    if args.dry_run:
        print("=== SUMMARY ===")
        print(_build_summary_payload(channel_id, entries, title, color))
        print("=== DETAILS ===")
        print(details_text)
        return 0

    if not webhook or not token or not server or not channel_id:
        print(
            "Error: webhook, bot token, server, and channel ID are required",
            file=sys.stderr,
        )
        return 2

    # Post summary via webhook
    summary_payload = _build_summary_payload(channel_id, entries, title, color)
    post_webhook(webhook, summary_payload)

    # Fetch last post ID for threading
    req = Request(
        f"{server.rstrip('/')}/api/v4/channels/{channel_id}/posts?page=0&per_page=1",
        headers={"Authorization": f"Bearer {token}"},
    )
    try:
        with urlopen(req, timeout=20) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            post_order = data.get("order", [])
            if not post_order:
                print("Error: no post found for threading", file=sys.stderr)
                return 2
            root_id = post_order[0]
    except (HTTPError, URLError) as e:
        print(f"Error fetching last post: {e}", file=sys.stderr)
        return 2

    _post_bot(server, token, channel_id, details_text, root_id=root_id)
    return 0
