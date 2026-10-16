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
from metrics.ingest import InFlightRunError, dump_records, ingest_run
from metrics.issues import GitHubIssueClient, open_issues
from metrics.logs import fetch_and_fingerprint, pending_jobs, stale_jobs
from metrics.report import build_catalogue, render_digest_markdown, write_reports
from metrics.rollup import derive_rollup, derive_rollups, open_issue_backlog
from metrics.router import apply_router
from metrics.rules import DEFAULT_RULES_PATH, classify_signatures, load_rules
from metrics.scrub import find_secrets
from metrics.store import dump_database, open_database, rebuild_database

REPO_ROOT = Path(__file__).parent.parent.parent
DEFAULT_DB = REPO_ROOT / "ci/metrics-data/metrics.db"
DEFAULT_SQL = REPO_ROOT / "ci/metrics-data/metrics.sql"
DEFAULT_REPORTS_DIR = REPO_ROOT / "ci/metrics-data/reports"
DEFAULT_SIGNATURES_DIR = REPO_ROOT / "ci/metrics-data/signatures"
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
    ingest.add_argument(
        "--skip-logs",
        action="store_true",
        help="Store run/job metadata only; leave log fingerprinting to fetch-logs.",
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
    rebuild.add_argument(
        "--init-if-missing",
        action="store_true",
        help="Start an empty database instead of failing when no dump exists yet "
        "(first run against a fresh ci-metrics branch).",
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

    rollup = metrics_sub.add_parser(
        "rollup",
        help="Derive CI failure metrics from stored facts and print JSON.",
    )
    rollup.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    rollup.add_argument(
        "--period",
        help="Data month to derive, e.g. 2026-10. Defaults to every period.",
    )
    rollup.set_defaults(func=cmd_rollup)

    report = metrics_sub.add_parser(
        "report",
        help="Write deterministic reports/latest.md and signatures/catalogue.md.",
    )
    report.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    report.add_argument(
        "--reports-dir",
        default=DEFAULT_REPORTS_DIR,
        type=Path,
        help=f"Report output directory (default: {DEFAULT_REPORTS_DIR}).",
    )
    report.add_argument(
        "--signatures-dir",
        default=DEFAULT_SIGNATURES_DIR,
        type=Path,
        help=f"Signature output directory (default: {DEFAULT_SIGNATURES_DIR}).",
    )
    report.set_defaults(func=cmd_report)

    digest = metrics_sub.add_parser(
        "digest",
        help="Render the Mattermost digest payload (attribution and delta first).",
    )
    digest.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    digest.add_argument(
        "--repo", default=DEFAULT_REPO, help="owner/repo the full report lives in."
    )
    digest.add_argument(
        "--report-ref",
        default="ci-metrics",
        help="Branch the full report is committed to (default: ci-metrics).",
    )
    digest.add_argument(
        "--report-path",
        default="ci/metrics-data/reports/latest.md",
        help="Path to the full report within --report-ref.",
    )
    digest.set_defaults(func=cmd_digest)

    open_issues_parser = metrics_sub.add_parser(
        "open-issues", help="Open or update GitHub issues for eligible CI failures."
    )
    open_issues_parser.add_argument(
        "--db",
        default=DEFAULT_DB,
        type=Path,
        help=f"SQLite database path (default: {DEFAULT_DB})",
    )
    open_issues_parser.add_argument(
        "--repo",
        default=DEFAULT_REPO,
        help="owner/repo whose issue tracker receives CI failure issues.",
    )
    open_issues_parser.add_argument(
        "--token",
        help="GitHub token used for issue creation; prefer BOT_TOKEN in CI.",
    )
    open_issues_parser.add_argument(
        "--issue-author",
        help="Expected login for adopted issues. Defaults to the token owner when posting.",
    )
    open_issues_parser.add_argument(
        "--repo-path",
        default=REPO_ROOT,
        type=Path,
        help="Git checkout used for freshness ancestry checks.",
    )
    open_issues_parser.add_argument(
        "--rules",
        default=DEFAULT_RULES_PATH,
        type=Path,
        help=f"Rule pack path used to add rule notes to issue bodies (default: {DEFAULT_RULES_PATH}).",
    )
    open_issues_parser.add_argument(
        "--min-occurrences",
        default=3,
        type=int,
        help="Minimum occurrences in a signature/substrate context.",
    )
    open_issues_parser.add_argument(
        "--min-runs",
        default=2,
        type=int,
        help="Minimum distinct runs in a signature/substrate context.",
    )
    open_issues_parser.add_argument(
        "--daily-budget",
        default=3,
        type=int,
        help="Maximum new issues to create per UTC day.",
    )
    open_issues_parser.add_argument(
        "--resolved-cooldown-days",
        default=14,
        type=int,
        help="Do not reopen issues closed as completed within this many days.",
    )
    open_issues_parser.add_argument(
        "--reopen-comment-window-days",
        default=7,
        type=int,
        help="Minimum days between unknown-freshness comments on a closed issue.",
    )
    open_issues_parser.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path to update after posting (default: {DEFAULT_SQL}).",
    )
    open_issues_parser.add_argument(
        "--no-dump",
        action="store_true",
        help="Do not update the deterministic SQL dump after posting.",
    )
    open_issues_parser.add_argument(
        "--dry-run",
        dest="dry_run",
        action="store_true",
        default=True,
        help="Preview actions without mutating GitHub or the issues table (default).",
    )
    open_issues_parser.add_argument(
        "--post",
        dest="dry_run",
        action="store_false",
        help="Create/update GitHub issues. Only use after the validation gate passes.",
    )
    open_issues_parser.add_argument(
        "--json", action="store_true", help="Print issue bridge summary as JSON."
    )
    open_issues_parser.set_defaults(func=cmd_open_issues)

    fetch_logs = metrics_sub.add_parser(
        "fetch-logs",
        help="Fetch and fingerprint logs for stored failed jobs not yet fingerprinted.",
    )
    _add_common_args(fetch_logs)
    fetch_logs.add_argument(
        "--run-id",
        type=int,
        action="append",
        help="Only consider jobs from this workflow run ID (repeatable).",
    )
    fetch_logs.add_argument(
        "--limit", type=int, help="Maximum number of pending jobs to consider."
    )
    fetch_logs.add_argument(
        "--reextract",
        action="store_true",
        help=(
            "Also re-fingerprint jobs signed by an older extractor/normaliser "
            "version, while GitHub still retains their logs."
        ),
    )
    fetch_logs.add_argument(
        "--json", action="store_true", help="Print the fetch summary as JSON."
    )
    fetch_logs.add_argument(
        "--sql",
        default=DEFAULT_SQL,
        type=Path,
        help=f"SQL dump path to update after fetching (default: {DEFAULT_SQL}).",
    )
    fetch_logs.add_argument(
        "--no-dump",
        action="store_true",
        help="Do not update the deterministic SQL dump after fetching.",
    )
    fetch_logs.set_defaults(func=cmd_fetch_logs)


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
            status="completed",
            limit=args.limit,
        )
        run_ids.extend(run["id"] for run in runs)

    summaries = []
    for run_id in run_ids:
        try:
            records = ingest_run(client, conn, run_id, fetch_logs=not args.skip_logs)
        except InFlightRunError as exc:
            # Belt-and-suspenders: the workflow-window path above already
            # requests status=completed, but a run can still be in-flight
            # when passed via --run-id, or finish between listing and
            # fetching. Skip it rather than failing the whole sweep — the
            # next ingest (hourly reconciliation or the next workflow_run
            # trigger) will pick it up once it is actually done.
            print(f"skipping run {run_id}: {exc}", file=sys.stderr)
            continue
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
    if not args.sql.exists() and args.init_if_missing:
        if args.db.exists():
            args.db.unlink()
        open_database(args.db, dump_path=args.sql).close()
        print(f"no dump at {args.sql}, initialised empty {args.db}")
        return 0
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


def cmd_rollup(args: argparse.Namespace) -> int:
    """Derive rollup metrics without storing aggregates."""

    conn = open_database(args.db)
    payload = derive_rollup(conn, args.period) if args.period else derive_rollups(conn)
    print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def cmd_report(args: argparse.Namespace) -> int:
    """Write deterministic human-facing report exports."""

    conn = open_database(args.db)
    latest_path, catalogue_path = write_reports(
        args.reports_dir,
        args.signatures_dir,
        derive_rollups(conn),
        build_catalogue(conn),
        open_issue_backlog(conn),
    )
    print(f"wrote {latest_path}")
    print(f"wrote {catalogue_path}")
    return 0


def cmd_digest(args: argparse.Namespace) -> int:
    """Print a Mattermost ``{"text": ...}`` payload for the daily digest."""

    conn = open_database(args.db)
    report_url = (
        f"https://github.com/{args.repo}/blob/{args.report_ref}/{args.report_path}"
    )
    text = render_digest_markdown(
        derive_rollups(conn), report_url, open_issue_backlog(conn)
    )
    print(json.dumps({"text": text}))
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


def cmd_open_issues(args: argparse.Namespace) -> int:
    """Open/update GitHub issues for recurring classified CI failures."""

    conn = open_database(args.db)
    pack = load_rules(args.rules)
    rule_notes = {rule.id: rule.notes for rule in pack.rules}
    if args.dry_run:
        client = (
            GitHubIssueClient(repo=args.repo, token=args.token) if args.token else None
        )
    else:
        client = GitHubIssueClient(repo=args.repo, token=args.token)
    if client is None:

        class DryRunClient:
            def current_user(self) -> str:
                return ""

            def list_issues(
                self, labels: list[str], state: str = "open", creator: str = ""
            ) -> list[dict]:
                return []

            def create_issue(self, title: str, body: str, labels: list[str]) -> dict:
                raise AssertionError("dry run must not create issues")

            def get_issue(self, number: int) -> dict:
                raise AssertionError("dry run must not fetch issues")

            def edit_issue(self, number: int, **fields: object) -> dict:
                raise AssertionError("dry run must not edit issues")

            def add_comment(self, number: int, body: str) -> dict:
                raise AssertionError("dry run must not comment")

            def closing_commit(self, number: int) -> Optional[str]:
                raise AssertionError("dry run must not sync issue state")

        client = DryRunClient()

    with conn:
        result = open_issues(
            conn,
            client,
            dry_run=args.dry_run,
            issue_author=args.issue_author,
            repo_path=args.repo_path,
            min_occurrences=args.min_occurrences,
            min_runs=args.min_runs,
            daily_budget=args.daily_budget,
            resolved_cooldown_days=args.resolved_cooldown_days,
            reopen_comment_window_days=args.reopen_comment_window_days,
            rule_notes=rule_notes,
        )
        if not args.dry_run and not args.no_dump:
            dump_database(conn, args.sql)

    summary = result.as_dict()
    if args.json:
        print(json.dumps(summary, indent=2, sort_keys=True))
    else:
        for key, value in summary.items():
            print(f"{key}: {len(value)}")
            for item in value:
                print("    - " + str(item))
    return 0


def cmd_fetch_logs(args: argparse.Namespace) -> int:
    """Fingerprint stored failed jobs whose logs have not been fetched yet."""

    client = GitHubClient(repo=args.repo, token=args.token)
    conn = open_database(args.db)
    jobs = pending_jobs(conn, run_ids=args.run_id, limit=args.limit)
    if args.reextract:
        remaining = None if args.limit is None else max(0, args.limit - len(jobs))
        if remaining != 0:
            jobs += stale_jobs(conn, run_ids=args.run_id, limit=remaining)
    summary = {"pending": len(jobs), **fetch_and_fingerprint(client, conn, jobs)}
    if args.json:
        print(json.dumps(summary, indent=2, sort_keys=True))
    else:
        print(
            "fetch-logs: {pending} pending, {fingerprinted} fingerprinted into "
            "{signatures} signatures ({reextracted} re-extracted), {log_missing} "
            "without log (runner-lost evidence), {log_gone} log gone, {routed} "
            "routed by metadata, {errors} errors".format(**summary)
        )
    if not args.no_dump:
        dump_database(conn, args.sql)
    return 0
