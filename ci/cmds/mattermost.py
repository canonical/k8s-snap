#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
Mattermost subcommands for `k8s-ci`.

This module provides functionality to post aggregated CI results or raw messages
to a Mattermost channel via incoming webhooks or bot accounts. It supports
posting a concise summary via webhook (with color support) and then posting
detailed results as a threaded comment using a bot.
"""

import argparse
import json
import os
import sys
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


def add_mattermost_cmds(parser: argparse.ArgumentParser) -> None:
    """
    Register Mattermost-related subcommands to the given CLI parser.

    Args:
        parser: The parent argparse.ArgumentParser to which subcommands will be added.
    """
    mattermost_parser = parser.add_parser(
        "mattermost", help="Post results or messages to Mattermost."
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

    # Post raw JSON command
    p = mattermost_sub.add_parser("post", help="Post raw JSON to a Mattermost channel.")
    p.add_argument(
        "--file", "-f", default=None, help="Message file (JSON) or '-' for stdin."
    )
    p.add_argument(
        "--webhook", "-w", help="Incoming webhook URL (or MATTERMOST_WEBHOOK_URL)"
    )
    p.add_argument("--dry-run", action="store_true", help="Print payload and exit")
    p.set_defaults(func=cmd_post)


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


def _build_tree_message(entries: List[Dict[str, Any]]) -> str:
    """
    Build a tree-formatted string of all test results.

    Args:
        entries: List of result dictionaries.

    Returns:
        Formatted tree string.
    """
    # Build a nested tree: channel -> os -> arch -> list of entries
    tree: Dict[str, Dict[str, Dict[str, List[Dict[str, Any]]]]] = {}
    for e in entries:
        ch = str(e.get("channel", "unknown"))
        osn = str(e.get("os", "unknown"))
        arch = str(e.get("arch", "unknown"))
        tree.setdefault(ch, {}).setdefault(osn, {}).setdefault(arch, []).append(e)

    lines: List[str] = []
    for ch in sorted(tree.keys(), reverse=True):
        lines.append(f"{ch}")
        os_list = sorted(tree[ch].keys())
        for oi, osn in enumerate(os_list):
            os_prefix = "└──" if oi == len(os_list) - 1 else "├──"
            lines.append(f"    {os_prefix} {osn}")
            arch_list = sorted(tree[ch][osn].keys())
            for ai, arch in enumerate(arch_list):
                arch_prefix = "└──" if ai == len(arch_list) - 1 else "├──"
                entry_list = tree[ch][osn][arch]
                failed_runs = [
                    e
                    for e in entry_list
                    if str(e.get("status", "")).lower() != "success"
                ]
                # Determine arch-level status: success only if all entries succeeded
                statuses = [str(e.get("status", "")).lower() for e in entry_list]
                status = (
                    "success" if all(s == "success" for s in statuses) else "failed"
                )
                emoji = ":white_check_mark:" if status == "success" else ":x:"
                # If there's only one entry, preserve previous behavior to link that run
                if not failed_runs:
                    run_link = (
                        _determine_run_link(entry_list[0]) if len(entry_list) else None
                    )
                    run_part = f" [Run]({run_link})" if run_link else " Run"
                else:
                    # For multiple entries, avoid a top-level link; individual failed runs below will have links.
                    run_part = ""
                indent = "        " if oi == len(os_list) - 1 else "    │   "
                lines.append(f"{indent}{arch_prefix} {arch}: {emoji} {run_part}")

                # If there are failed runs for this arch, add a subtree listing them with links
                if failed_runs:
                    # child prefix keeps vertical bar if this arch isn't the last in the os list
                    child_base = indent + (
                        "    " if ai == len(arch_list) - 1 else "│   "
                    )
                    for fi, fe in enumerate(failed_runs):
                        child_conn = "└──" if fi == len(failed_runs) - 1 else "├──"
                        run_link = _determine_run_link(fe)
                        run_part = f" [Run]({run_link})" if run_link else " Run"
                        name = (
                            fe.get("test").split("::")[-1]
                            if fe.get("test")
                            else "Unnamed Test"
                        )
                        lines.append(f"{child_base}{child_conn} {name}{run_part}")
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
    successes = sum(1 for e in entries if str(e.get("status", "")).lower() == "success")
    skipped = sum(1 for e in entries if str(e.get("status", "")).lower() == "skipped")
    failures = total - successes - skipped
    pass_pct = (successes / total * 100) if total else 0.0
    emoji = ":white_check_mark:" if failures == 0 else ":x:"
    title = f"{emoji} {title}"
    text = f"{successes}/{total} passed ({pass_pct:.0f}%) — Skipped: {skipped}"
    return {
        "channel_id": channel_id,
        "attachments": [
            {"fallback": text, "title": title, "text": text, "color": color}
        ],
    }


def _determine_color(entries: List[Dict[str, Any]], text: str) -> str:
    """
    Determine color for webhook attachment based on test statuses.

    Returns:
        'good' if all success, 'danger' if any failed.
    """
    for e in entries:
        if str(e.get("status", "")).lower() != "success":
            return "danger"
    if ":x:" in text:
        return "danger"
    return "good"


def _post_webhook(webhook: str, payload: Dict[str, Any]) -> None:
    """Post payload to Mattermost using an incoming webhook."""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = Request(
        webhook, data=body, headers={"Content-Type": "application/json; charset=utf-8"}
    )
    try:
        with urlopen(req, timeout=20) as resp:
            resp.read()  # webhook returns plain "ok"
    except (HTTPError, URLError) as e:
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
    except (HTTPError, URLError) as e:
        print(f"Bot post error: {e}", file=sys.stderr)
        raise SystemExit(2)


def cmd_results_message(args: argparse.Namespace) -> int:
    """Command to post CI results summary via webhook and details as bot-threaded comment."""
    entries = _load_flattened_json(args.file)
    tree_text = _build_tree_message(entries)
    color = _determine_color(entries, tree_text)

    webhook = args.webhook or os.environ.get("MATTERMOST_WEBHOOK_URL")
    token = args.bot_token or os.environ.get("MATTERMOST_BOT_TOKEN")
    server = args.server or os.environ.get("MATTERMOST_SERVER")
    channel_id = args.channel_id or os.environ.get("MATTERMOST_CHANNEL_ID")
    title = args.title or "CI results"

    if args.dry_run:
        print("=== SUMMARY ===")
        print(_build_summary_payload(channel_id, entries, title, color))
        print("=== TREE ===")
        print(tree_text)
        return 0

    if not webhook or not token or not server or not channel_id:
        print(
            "Error: webhook, bot token, server, and channel ID are required",
            file=sys.stderr,
        )
        return 2

    # Post summary via webhook
    summary_payload = _build_summary_payload(channel_id, entries, title, color)
    _post_webhook(webhook, summary_payload)

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

    # Post detailed tree as threaded comment
    _post_bot(server, token, channel_id, tree_text, root_id=root_id)
    return 0


def cmd_post(args: argparse.Namespace) -> int:
    """Post a raw JSON message either via webhook or bot account."""
    if args.file == "-":
        payload = json.load(sys.stdin)
    else:
        with open(args.file, "r", encoding="utf-8") as fh:
            payload = json.load(fh)

    webhook = args.webhook or os.environ.get("MATTERMOST_WEBHOOK_URL")
    token = args.bot_token or os.environ.get("MATTERMOST_BOT_TOKEN")
    server = args.server or os.environ.get("MATTERMOST_SERVER")
    channel_id = args.channel_id or os.environ.get("MATTERMOST_CHANNEL_ID")

    if args.dry_run:
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        return 0

    if webhook:
        _post_webhook(webhook, payload)
    elif token:
        _post_bot(server, token, channel_id, json.dumps(payload, ensure_ascii=False))
    else:
        print("Error: webhook or bot token required", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_mattermost_cmds(parser)
    args = parser.parse_args()
    sys.exit(args.func(args))
