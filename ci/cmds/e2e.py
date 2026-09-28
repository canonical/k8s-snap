#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""
e2e-tests.yaml subcommands for `k8s-ci`.

This module turns the run context collected by the "Collect run context"
step of `.github/workflows/e2e-tests.yaml` into the human-readable
failure-context table and the machine-readable `result.json`, so the
formatting logic lives in one place instead of being duplicated across bash
heredocs.
"""

import argparse
import json
import os
from typing import Any, Dict

# Fields collected by the workflow's "Collect run context" step, exported to
# GITHUB_ENV as ctx_<field>, and therefore already present as environment
# variables by the time these commands run.
_CTX_FIELDS = (
    "test",
    "channel",
    "revision",
    "artifact",
    "os",
    "arch",
    "substrate",
    "flavor",
    "runner",
    "attempt",
    "sha",
    "source",
    "run_url",
)


def add_e2e_cmds(parser: argparse.ArgumentParser) -> None:
    """
    Register e2e-tests.yaml-related subcommands to the given CLI parser.

    Args:
        parser: The parent argparse.ArgumentParser to which subcommands will be added.
    """
    e2e_parser = parser.add_parser("e2e", help="e2e-tests.yaml CI helpers.")
    e2e_sub = e2e_parser.add_subparsers(
        dest="e2e_command", required=True, title="e2e commands"
    )

    p = e2e_sub.add_parser(
        "print-failure-context",
        help="Print a markdown table describing a failed e2e job.",
    )
    p.set_defaults(func=cmd_print_failure_context)

    p = e2e_sub.add_parser(
        "write-result-json",
        help="Write the machine-readable result.json for an e2e job.",
    )
    p.add_argument(
        "--output", "-o", default="results/result.json", help="Output file path."
    )
    p.set_defaults(func=cmd_write_result_json)


def _context_from_env() -> Dict[str, str]:
    """Collect the ctx_* fields set by the 'Collect run context' workflow step."""
    return {field: os.environ.get(f"ctx_{field}", "") for field in _CTX_FIELDS}


def build_failure_context_table(ctx: Dict[str, str], test_name: str) -> str:
    """
    Render the run context as a markdown table for the job log and summary.

    Args:
        ctx: Context fields collected by the workflow (see _context_from_env).
        test_name: Sanitized name used for the inspection report artifact.

    Returns:
        Markdown block, ending in a trailing blank line.
    """
    channel = ctx["channel"] or f"n/a (artifact: {ctx['artifact']})"
    revision = ctx["revision"] or "n/a (locally built snap)"
    test = ctx["test"]

    rows = (
        ("Test", f"`{test}`"),
        ("Channel", channel),
        ("Revision", revision),
        ("OS", ctx["os"]),
        ("Arch", ctx["arch"]),
        ("Substrate", ctx["substrate"]),
        ("Flavor", ctx["flavor"]),
        ("Runner", ctx["runner"]),
        ("Attempt", ctx["attempt"]),
        ("SHA", f"`{ctx['sha']}`"),
        ("Test source", ctx["source"]),
        ("Run", ctx["run_url"]),
        ("Artifacts", f"inspection-reports-{test_name}"),
    )

    lines = ["### :x: " + test, "", "| Field | Value |", "| --- | --- |"]
    lines += [f"| {field} | {value} |" for field, value in rows]
    lines.append("")
    return "\n".join(lines)


def cmd_print_failure_context(args: argparse.Namespace) -> int:
    """Print the failure-context table for the current job to stdout."""
    ctx = _context_from_env()
    test_name = os.environ.get("test_name", "")
    print(build_failure_context_table(ctx, test_name))
    return 0


def build_result(ctx: Dict[str, str], status: str) -> Dict[str, Any]:
    """
    Build the machine-readable result dict written to result.json.

    `channel`, `revision`, and `artifact` are kept raw (empty for fields that
    don't apply to the run) rather than the human-readable fallbacks used in
    the failure table, so existing consumers that group by channel/os/arch
    (e.g. `ci/cmds/mattermost.py`) are unaffected.
    """
    return {
        "os": ctx["os"],
        "arch": ctx["arch"],
        "channel": ctx["channel"],
        "revision": ctx["revision"],
        "artifact": ctx["artifact"],
        "substrate": ctx["substrate"],
        "flavor": ctx["flavor"],
        "runner": ctx["runner"],
        "attempt": ctx["attempt"],
        "sha": ctx["sha"],
        "source": ctx["source"],
        "run_url": ctx["run_url"],
        "test": ctx["test"],
        "status": status,
    }


def cmd_write_result_json(args: argparse.Namespace) -> int:
    """Write result.json for the current job and echo it to stdout."""
    ctx = _context_from_env()
    status = os.environ.get("JOB_STATUS", "")
    result = build_result(ctx, status)

    out_dir = os.path.dirname(args.output)
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)
    with open(args.output, "w", encoding="utf-8") as fh:
        json.dump(result, fh)

    print(json.dumps(result, indent=2))
    return 0
