#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""Metrics subcommands for `k8s-ci`."""

from __future__ import annotations

import argparse
import json
import re
import sqlite3
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Optional

from metrics.gh import DEFAULT_REPO, GitHubClient
from metrics.ingest import dump_records, ingest_run
from metrics.router import apply_router
from metrics.rules import DEFAULT_RULES_PATH, classify_signatures, load_rules
from metrics.scrub import find_secrets
from metrics.store import dump_database, open_database, rebuild_database

REPO_ROOT = Path(__file__).parent.parent.parent
DEFAULT_DB = REPO_ROOT / "ci/metrics-data/metrics.db"
DEFAULT_SQL = REPO_ROOT / "ci/metrics-data/metrics.sql"
WORKFLOW_ALIASES = {
    "nightly": "nightly-test.yaml",
    "weekly": "weekly-test.yaml",
    "pr": "lint_and_integration.yaml",
    "lint": "lint_and_integration.yaml",
}
SINCE_RE = re.compile(r"^(\d+)([dwm])$")


def add_metrics_cmds(parser: argparse.ArgumentParser) -> None:
    """Register metrics-related subcommands."""

    metrics_parser = parser.add_parser(
        "metrics", help="CI failure classification and metrics."
    )
    metrics_sub = metrics_parser.add_subparsers(
        dest="metrics_command", required=True, title="metrics commands"
    )

    ingest = metrics_sub.add_parser(
        "ingest", help="Ingest completed workflow run attempts into SQLite."
    )
    _add_common_args(ingest)
    ingest.add_argument(
        "--run-id", type=int, action="append", help="Workflow run ID (repeatable)."
    )
    ingest.add_argument(
        "--workflow",
        help=f"Workflow to ingest. Aliases: {', '.join(sorted(WORKFLOW_ALIASES))}.",
    )
    ingest.add_argument("--since", help="Backfill window, e.g. 90d, 12w, 3m.")
    ingest.add_argument("--limit", type=int, help="Maximum number of runs to ingest.")
    ingest.add_argument(
        "--json", action="store_true", help="Print ingested attempt summaries as JSON."
    )
    ingest.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path to update after ingest (default: {DEFAULT_SQL}).",
    )
    ingest.add_argument(
        "--no-dump",
        action="store_true",
        help="Do not update the deterministic SQL dump after ingest.",
    )
    ingest.set_defaults(func=cmd_ingest)

    audit = metrics_sub.add_parser(
        "audit-secrets",
        help="Fail if stored excerpts or selected files contain unredacted secrets.",
    )
    audit.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path to audit when no paths are given (default: {DEFAULT_DB})",
    )
    audit.add_argument(
        "paths",
        nargs="*",
        type=Path,
        help="Files or directories to audit instead of the metrics database.",
    )
    audit.set_defaults(func=cmd_audit_secrets)

    dump = metrics_sub.add_parser(
        "dump", help="Write deterministic durable-table SQL dump."
    )
    dump.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    dump.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path (default: {DEFAULT_SQL})",
    )
    dump.set_defaults(func=cmd_dump)

    rebuild = metrics_sub.add_parser(
        "rebuild", help="Rebuild metrics.db from the deterministic SQL dump."
    )
    rebuild.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    rebuild.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path (default: {DEFAULT_SQL})",
    )
    rebuild.set_defaults(func=cmd_rebuild)

    classify = metrics_sub.add_parser(
        "classify", help="Apply Stage A router and Stage B rules to stored failures."
    )
    classify.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    classify.add_argument(
        "--rules",
        default=DEFAULT_RULES_PATH,
        type=Path,
        help=f"Rule pack path (default: {DEFAULT_RULES_PATH})",
    )
    classify.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path to update after classification (default: {DEFAULT_SQL}).",
    )
    classify.add_argument(
        "--no-dump",
        action="store_true",
        help="Do not update the deterministic SQL dump after classification.",
    )
    classify.add_argument(
        "--json", action="store_true", help="Print classification summary as JSON."
    )
    classify.set_defaults(func=cmd_classify)

    for name in ("fetch-logs", "rollup", "report", "open-issues"):
        stub = metrics_sub.add_parser(
            name, help="Stub for a later CI failure classification PR."
        )
        stub.set_defaults(func=_cmd_stub)


def _add_common_args(parser: argparse.ArgumentParser) -> None:
    parser.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    parser.add_argument("--repo", default=DEFAULT_REPO, help="owner/repo to query.")
    parser.add_argument("--token", help="GitHub token with actions:read.")


def _parse_since(value: str) -> str:
    match = SINCE_RE.match(value)
    if not match:
        raise argparse.ArgumentTypeError("expected a window like 90d, 12w, or 3m")
    amount = int(match.group(1))
    unit = match.group(2)
    days = amount if unit == "d" else amount * 7 if unit == "w" else amount * 30
    start = datetime.now(timezone.utc) - timedelta(days=days)
    return f">={start.strftime('%Y-%m-%d')}"


def _workflow_ref(value: Optional[str]) -> Optional[str]:
    if value is None:
        return None
    return WORKFLOW_ALIASES.get(value, value)


def cmd_ingest(args: argparse.Namespace) -> int:
    """Ingest run attempts selected by run ID or workflow window."""

    if not args.run_id and not args.workflow:
        print("metrics ingest requires --run-id or --workflow", file=sys.stderr)
        return 2
    if args.workflow and not args.since and not args.run_id:
        print("metrics ingest --workflow requires --since", file=sys.stderr)
        return 2

    client = GitHubClient(repo=args.repo, token=args.token)
    conn = open_database(args.db)
    run_ids = list(args.run_id or [])

    if args.workflow:
        workflow = _workflow_ref(args.workflow)
        runs = client.list_workflow_runs(
            workflow,
            created=_parse_since(args.since) if args.since else None,
            limit=args.limit,
        )
        run_ids.extend(run["id"] for run in runs)

    summaries = []
    for run_id in run_ids:
        records = ingest_run(client, conn, run_id)
        summaries.extend(dump_records(records))

    if args.json:
        print(json.dumps(summaries, indent=2, sort_keys=True))
    else:
        for summary in summaries:
            print(
                "ingested run {run_id} attempt {attempt}: {jobs_failure} failed jobs, "
                "{retry_outcomes} retry outcomes".format(**summary)
            )
    if not args.no_dump:
        dump_database(conn, args.sql)
    return 0


def cmd_dump(args: argparse.Namespace) -> int:
    conn = open_database(args.db, dump_path=args.sql)
    dump_database(conn, args.sql)
    print(f"wrote {args.sql}")
    return 0


def cmd_rebuild(args: argparse.Namespace) -> int:
    rebuild_database(args.db, args.sql).close()
    print(f"rebuilt {args.db} from {args.sql}")
    return 0


def cmd_classify(args: argparse.Namespace) -> int:
    """Apply metadata routing and the ordered Stage B rule pack."""

    conn = open_database(args.db)
    pack = load_rules(args.rules)
    with conn:
        router_summary = apply_router(conn)
        rules_summary = classify_signatures(conn, pack)

    summary = {
        "router": router_summary,
        "rules": {"ruleset_version": pack.version, **rules_summary},
    }
    if args.json:
        print(json.dumps(summary, indent=2, sort_keys=True))
    else:
        print("router: {routed} routed, {deferred} deferred".format(**router_summary))
        print(
            "rules v{ruleset_version}: {matched} matched, {unmatched} unmatched".format(
                **summary["rules"]
            )
        )

    if not args.no_dump:
        dump_database(conn, args.sql)
    return 0


def _iter_files(paths: list[Path]) -> list[Path]:
    files = []
    for path in paths:
        if path.is_dir():
            files.extend(child for child in sorted(path.rglob("*")) if child.is_file())
        else:
            files.append(path)
    return files


def _audit_text(label: str, text: str) -> int:
    findings = find_secrets(text)
    for finding in findings:
        message = (
            f"{label}:{finding.line_number}: secret-like value ({finding.rule_name})"
        )
        print(message, file=sys.stderr)
    return len(findings)


def _audit_database(db_path: Path) -> int:
    if not db_path.exists():
        print(f"metrics audit-secrets: database not found: {db_path}", file=sys.stderr)
        return -1

    conn = sqlite3.connect(db_path)
    tables = {
        row[0]
        for row in conn.execute("SELECT name FROM sqlite_master WHERE type = 'table'")
    }
    findings = 0
    if "job_excerpts" in tables:
        rows = conn.execute("SELECT job_id, excerpt FROM job_excerpts ORDER BY job_id")
        for job_id, excerpt in rows:
            findings += _audit_text(f"job_excerpts:{job_id}", excerpt)
    if "signatures" in tables:
        rows = conn.execute(
            """
            SELECT signature_id, normalised_pattern, exemplar_excerpt
            FROM signatures ORDER BY signature_id
            """
        )
        for signature_id, normalised_pattern, exemplar_excerpt in rows:
            findings += _audit_text(
                f"signatures:{signature_id}:normalised", normalised_pattern
            )
            findings += _audit_text(
                f"signatures:{signature_id}:excerpt", exemplar_excerpt
            )
    return findings


def cmd_audit_secrets(args: argparse.Namespace) -> int:
    """Audit stored or soon-to-be-stored text for unredacted secrets."""

    findings = 0
    if args.paths:
        for path in _iter_files(args.paths):
            try:
                text = path.read_text(encoding="utf-8")
            except UnicodeDecodeError as exc:
                print(
                    f"metrics audit-secrets: cannot decode {path}: {exc}",
                    file=sys.stderr,
                )
                return 2
            findings += _audit_text(str(path), text)
    else:
        findings = _audit_database(args.db)
        if findings < 0:
            return 2

    if findings:
        print(f"metrics audit-secrets: {findings} finding(s)", file=sys.stderr)
        return 1
    print("metrics audit-secrets: no findings")
    return 0


def _cmd_stub(args: argparse.Namespace) -> int:
    print(
        f"metrics {args.metrics_command} is intentionally left for a later PR.",
        file=sys.stderr,
    )
    return 2
