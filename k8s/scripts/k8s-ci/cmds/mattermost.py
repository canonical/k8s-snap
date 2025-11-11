#!/usr/bin/env python3
#
# Copyright 2025 Canonical, Ltd.
#

import argparse
import json
import os
import sys
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


def add_mattermost_cmds(parser: argparse.ArgumentParser) -> None:
    mattermost_parser = parser.add_parser(
        "mattermost", help="Post results or messages to Mattermost."
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
        "post", help="Post a raw JSON message to a Mattermost channel."
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
        flat = []
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
    tree = {}
    for e in entries:
        ch = str(e.get("channel", "unknown"))
        osn = str(e.get("os", "unknown"))
        arch = str(e.get("arch", "unknown"))
        tree.setdefault(ch, {}).setdefault(osn, {})[arch] = e

    lines = []
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


def _post_webhook(webhook: str, payload: Dict[str, Any]) -> Dict[str, Any]:
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = Request(
        webhook, data=body, headers={"Content-Type": "application/json; charset=utf-8"}
    )
    try:
        with urlopen(req, timeout=20) as resp:
            resp_body = resp.read().decode("utf-8", errors="ignore")
            if resp_body.strip():
                return json.loads(resp_body)
            return {}
    except HTTPError as e:
        print(
            f"HTTP error: {e.code} {e.read().decode('utf-8', errors='ignore')}",
            file=sys.stderr,
        )
        raise SystemExit(2)
    except URLError as e:
        print(f"Network error: {e.reason}", file=sys.stderr)
        raise SystemExit(2)


def _post_comment(server: str, token: str, root_id: str, text: str) -> None:
    body = {"message": text, "root_id": root_id}
    req = Request(
        f"{server}/api/v4/posts",
        data=json.dumps(body).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
    )
    try:
        with urlopen(req, timeout=20) as resp:
            resp.read()
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
    tree_text = _build_tree_message(entries)

    summary = "Results summary available. See thread for details."
    title = args.title or ""
    color = _determine_color(entries, tree_text)

    payload = _build_payload(summary, title.strip(), color)

    webhook = (
        args.webhook
        or os.environ.get("MATTERMOST_WEBHOOK_URL")
        or os.environ.get("MATTERMOST_BOT_WEBHOOK_URL")
    )
    if not webhook:
        print(
            "Error: webhook required via --webhook or MATTERMOST_WEBHOOK_URL",
            file=sys.stderr,
        )
        return 2

    token = os.environ.get("MATTERMOST_BOT_TOKEN")
    server = os.environ.get("MATTERMOST_SERVER")
    if not token or not server:
        print(
            "Error: MATTERMOST_BOT_TOKEN and MATTERMOST_SERVER required for threaded comment",
            file=sys.stderr,
        )
        return 2

    if args.dry_run:
        print("=== SUMMARY PAYLOAD ===")
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        print("=== TREE ===")
        print(tree_text)
        return 0

    resp = _post_webhook(webhook, payload)
    root_id = resp.get("id")
    if not root_id:
        print("Error: webhook response missing post id", file=sys.stderr)
        return 2

    _post_comment(server, token, root_id, tree_text)
    return 0


def cmd_post(args: argparse.Namespace) -> int:
    if args.file == "-":
        payload = json.load(sys.stdin)
    else:
        with open(args.file, "r", encoding="utf-8") as fh:
            payload = json.load(fh)

    webhook = (
        args.webhook
        or os.environ.get("MATTERMOST_WEBHOOK_URL")
        or os.environ.get("MATTERMOST_BOT_WEBHOOK_URL")
    )
    if not webhook:
        print(
            "Error: webhook required via --webhook or MATTERMOST_WEBHOOK_URL",
            file=sys.stderr,
        )
        return 2

    if args.dry_run:
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        return 0

    _post_webhook(webhook, payload)
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    add_mattermost_cmds(parser)
    args = parser.parse_args()
    sys.exit(args.func(args))
