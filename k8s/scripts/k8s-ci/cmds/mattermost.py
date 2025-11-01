#!/usr/bin/env python3
#
# Copyright 2025 Canonical, Ltd.
#
"""
Subcommand implementation for `k8s-ci mattermost`.
This module exposes `add_mattermost_cmds(subparsers)` which
registers mattermost subcommands to the passed cli parser.
"""
import argparse
import json
import os
import sys
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


def add_mattermost_cmds(parser: argparse.ArgumentParser) -> None:
    """Register the `mattermost` subcommand and its actions."""
    mattermost_parser = parser.add_parser(
        "mattermost",
        help="Post results or messages to Mattermost.",
    )
    mattermost_sub = mattermost_parser.add_subparsers(
        dest="mattermost_command", required=True, title="mattermost commands"
    )

    p = mattermost_sub.add_parser(
        "results-message",
        help="Aggregate a list of results to a Mattermost message and post it.",
    )
    p.add_argument(
        "--file", "-f", default=None, help="results file (json) or '-' for stdin."
    )
    p.add_argument(
        "--webhook",
        "-w",
        help="Mattermost incoming webhook URL (or set MATTERMOST_WEBHOOK_URL)",
    )
    p.add_argument("--title", "-t", default=None, help="message title")
    p.add_argument("--dry-run", action="store_true", help="print payload and exit")
    p.set_defaults(func=cmd_results_message)

    p = mattermost_sub.add_parser(
        "post",
        help="Post a raw JSON message to a Mattermost channel.",
    )
    p.add_argument(
        "--file", "-f", default=None, help="message file (json) or '-' for stdin."
    )
    p.add_argument(
        "--webhook",
        "-w",
        help="Mattermost incoming webhook URL (or set MATTERMOST_WEBHOOK_URL)",
    )
    p.add_argument("--dry-run", action="store_true", help="print payload and exit")
    p.set_defaults(func=cmd_post)


def _load_flattened_json(path: Optional[str]) -> List[Dict[str, Any]]:
    if path == "-":
        data = json.load(sys.stdin)
    elif path:
        with open(path, "r", encoding="utf-8") as fh:
            data = json.load(fh)
    else:
        raise SystemExit("Error: must provide --file or '-' for stdin")

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
    run_url = entry.get("run_url") or entry.get("runLink") or entry.get("run_link")
    if run_url:
        return str(run_url)
    repo = os.environ.get("GITHUB_REPOSITORY")
    run_id = os.environ.get("GITHUB_RUN_ID")
    if repo and run_id:
        return f"https://github.com/{repo}/actions/runs/{run_id}"
    return None


def _build_tree_message(entries: List[Dict[str, Any]]) -> str:
    tree: Dict[str, Dict[str, Dict[str, Dict[str, Any]]]] = {}
    for e in entries:
        ch = str(e.get("channel", "unknown"))
        osn = str(e.get("os", "unknown"))
        arch = str(e.get("arch", "unknown"))
        tree.setdefault(ch, {}).setdefault(osn, {})[arch] = e

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
                entry = tree[ch][osn][arch]
                status = str(entry.get("status", "")).lower()
                emoji = ":white_check_mark:" if status == "success" else ":x:"
                label = "Succeeded" if status == "success" else "Failed"
                run_link = _determine_run_link(entry)
                indent = "        " if oi == len(os_list) - 1 else "    │   "
                run_part = f" [Run]({run_link})" if run_link else " Run"
                lines.append(f"{indent}{arch_prefix} {arch}: {emoji}{label} {run_part}")
    return "\n".join(lines)


def _determine_color(entries: List[Dict[str, Any]], text: str) -> str:
    for e in entries:
        if str(e.get("status", "")).lower() != "success":
            return "danger"
    if ":x:" in text:
        return "danger"
    return "good"


def _build_payload(
    text: str, title: str, color: str, fallback: str = ""
) -> Dict[str, Any]:
    return {
        "attachments": [
            {
                "fallback": fallback or title,
                "color": color,
                "title": title,
                "text": text,
            }
        ]
    }


def _post_to_mattermost(webhook: str, payload: Dict[str, Any]) -> None:
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = Request(
        webhook, data=body, headers={"Content-Type": "application/json; charset=utf-8"}
    )
    try:
        with urlopen(req, timeout=20) as resp:
            print(f"Posted payload: HTTP {resp.getcode()}")
            body = resp.read().decode("utf-8", errors="ignore")
            if body:
                print(body)
    except HTTPError as e:
        print(
            f"HTTP error: {e.code} {e.read().decode('utf-8', errors='ignore')}",
            file=sys.stderr,
        )
        raise SystemExit(2)
    except URLError as e:
        print(f"Network error: {e.reason}", file=sys.stderr)
        raise SystemExit(2)


def cmd_results_message(args: argparse.Namespace) -> int:
    entries = _load_flattened_json(args.file)
    message = _build_tree_message(entries)
    title = args.title or ""
    color = _determine_color(entries, message)
    payload = _build_payload(message, title.strip(), color)
    webhook = (
        args.webhook
        or os.environ.get("MATTERMOST_WEBHOOK_URL")
        or os.environ.get("MATTERMOST_BOT_WEBHOOK_URL")
    )
    if args.dry_run:
        print("=== MESSAGE ===")
        print(message)
        print("=== PAYLOAD ===")
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        return 0
    if not webhook:
        print(
            "Error: webhook required via --webhook or MATTERMOST_WEBHOOK_URL",
            file=sys.stderr,
        )
        return 2
    _post_to_mattermost(webhook, payload)
    return 0


def cmd_post(args: argparse.Namespace) -> int:
    payload = json.load(sys.stdin if args.file == "-" else open(args.file))
    webhook = (
        args.webhook
        or os.environ.get("MATTERMOST_WEBHOOK_URL")
        or os.environ.get("MATTERMOST_BOT_WEBHOOK_URL")
    )
    if args.dry_run:
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        return 0
    if not webhook:
        print(
            "Error: webhook required via --webhook or MATTERMOST_WEBHOOK_URL",
            file=sys.stderr,
        )
        return 2
    _post_to_mattermost(webhook, payload)
    return 0
