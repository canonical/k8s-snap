#!/usr/bin/env python3
#
# Copyright 2026 Canonical, Ltd.
#
"""Metrics subcommands for `k8s-ci`.

Provides CI failure classification and measurement:

    k8s-ci metrics ingest --run-id 34172139128
    k8s-ci metrics ingest --workflow nightly --since 90d
    k8s-ci metrics show --run-id 34172139128

See `docs/proposals/003-ci-failure-classification-and-metrics.md`.
"""

import argparse
import collections
import json
import logging
import os
import re
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional

from metrics import TAXONOMY_VERSION
from metrics.gh import DEFAULT_REPO, GitHubClient
from metrics.ingest import ingest_run, record_month, workflow_slug
from metrics.models import ClassifiedBy, FailureClass, JobFailure, StepRecord
from metrics.report import render_markdown, render_mattermost
from metrics.rollup import aggregate
from metrics.router import ROUTER_VERSION, apply_router, route, route_all
from metrics.rules import classify_with_rules, load_rules
from metrics.scrub import find_secrets
from metrics.signature import NORMALISER_VERSION, fingerprint
from metrics.store import DATA_ROOT, MetricsStore

LOG = logging.getLogger("k8s-ci.metrics")

REPO_ROOT = Path(__file__).parent.parent.parent

# Friendly aliases so operators do not have to memorise workflow file names.
WORKFLOW_ALIASES = {
    "nightly": "nightly-test.yaml",
    "weekly": "weekly-test.yaml",
    "pr": "lint_and_integration.yaml",
    "lint": "lint_and_integration.yaml",
}

SINCE_RE = re.compile(r"^(\d+)([dwm])$")


def add_metrics_cmds(parser: argparse.ArgumentParser) -> None:
    """Register metrics-related subcommands to the given CLI parser."""
    metrics_parser = parser.add_parser(
        "metrics", help="CI failure classification and metrics."
    )
    metrics_sub = metrics_parser.add_subparsers(
        dest="metrics_command", required=True, title="metrics commands"
    )

    ingest = metrics_sub.add_parser(
        "ingest", help="Ingest workflow run metadata into the metrics store."
    )
    _add_common_args(ingest)
    ingest.add_argument(
        "--run-id", type=int, action="append", help="Run ID (repeatable)."
    )
    ingest.add_argument(
        "--workflow",
        help=f"Workflow to ingest. Aliases: {', '.join(sorted(WORKFLOW_ALIASES))}.",
    )
    ingest.add_argument(
        "--since", help="Backfill window, e.g. 90d, 12w, 3m. Requires --workflow."
    )
    ingest.add_argument("--limit", type=int, help="Maximum number of runs to ingest.")
    ingest.add_argument(
        "--force",
        action="store_true",
        help="Re-ingest runs already present in the store.",
    )
    ingest.add_argument(
        "--no-flake-detection",
        action="store_true",
        help="Skip cross-attempt lookups (faster, loses flake ground truth).",
    )
    ingest.set_defaults(func=cmd_ingest)

    show = metrics_sub.add_parser("show", help="Print a stored run record summary.")
    _add_common_args(show)
    show.add_argument("--run-id", type=int, required=True, help="Run ID to display.")
    show.add_argument("--attempt", type=int, default=None, help="Run attempt.")
    show.add_argument(
        "--json", action="store_true", help="Emit the raw record as JSON."
    )
    show.set_defaults(func=cmd_show)

    classify = metrics_sub.add_parser(
        "classify",
        help="Classify stored failures (Stage A metadata router).",
    )
    _add_common_args(classify)
    classify.add_argument(
        "--run-id", type=int, action="append", help="Run ID (repeatable)."
    )
    classify.add_argument(
        "--workflow", help="Classify every stored run of this workflow."
    )
    classify.add_argument(
        "--job-id", type=int, help="Explain the verdict for a single job."
    )
    classify.add_argument(
        "--dry-run",
        action="store_true",
        help="Report verdicts without writing them back to the store.",
    )
    classify.add_argument(
        "--rules",
        help="Path to the rule pack (default: ci/failure_rules.yaml).",
    )
    classify.set_defaults(func=cmd_classify)

    fetch = metrics_sub.add_parser(
        "fetch-logs",
        help="Fetch logs for deferred failures and compute signatures (Tier 1).",
    )
    _add_common_args(fetch)
    fetch.add_argument(
        "--run-id", type=int, action="append", help="Run ID (repeatable)."
    )
    fetch.add_argument("--workflow", help="Process every stored run of this workflow.")
    fetch.add_argument(
        "--workers",
        type=int,
        default=12,
        help="Concurrent log downloads (default: %(default)s).",
    )
    fetch.add_argument(
        "--force",
        action="store_true",
        help="Refetch logs for failures that already carry a signature.",
    )
    fetch.set_defaults(func=cmd_fetch_logs)

    audit = metrics_sub.add_parser(
        "audit-secrets",
        help="Scan every stored excerpt for secrets (release gate).",
    )
    _add_common_args(audit)
    audit.set_defaults(func=cmd_audit_secrets)

    rollup = metrics_sub.add_parser(
        "rollup", help="Aggregate classified runs into a periodic rollup."
    )
    _add_common_args(rollup)
    rollup.add_argument(
        "--period",
        help="Period label to aggregate, e.g. 2026-09. Defaults to every "
        "month present in the store.",
    )
    rollup.add_argument("--workflow", help="Restrict to one workflow (alias or slug).")
    rollup.add_argument("--dry-run", action="store_true", help="Print without writing.")
    rollup.set_defaults(func=cmd_rollup)

    report = metrics_sub.add_parser(
        "report", help="Render a report from stored rollups."
    )
    _add_common_args(report)
    report.add_argument(
        "--format",
        choices=("mattermost", "markdown"),
        default="mattermost",
        help="Output format.",
    )
    report.add_argument(
        "--period", help="Period to report on. Defaults to the newest rollup."
    )
    report.add_argument(
        "--periods",
        type=int,
        default=6,
        help="How many periods the markdown trend table covers.",
    )
    report.add_argument("--out", help="Write to this path instead of stdout.")
    report.add_argument(
        "--post",
        action="store_true",
        help="Post to Mattermost. Without it, the report only goes to stdout.",
    )
    report.add_argument(
        "--webhook", help="Incoming webhook URL (or MATTERMOST_WEBHOOK_URL)."
    )
    report.set_defaults(func=cmd_report)

    reclass = metrics_sub.add_parser(
        "reclassify",
        help="Replay the rule pack over stored excerpts without refetching.",
    )
    _add_common_args(reclass)
    reclass.add_argument("--workflow", help="Restrict to one workflow (alias or slug).")
    reclass.add_argument("--period", help="Restrict to one period, e.g. 2026-09.")
    reclass.add_argument(
        "--rules", help="Path to the rule pack. Defaults to ci/failure_rules.yaml."
    )
    reclass.add_argument(
        "--dry-run",
        action="store_true",
        help="Report what would change without writing.",
    )
    reclass.set_defaults(func=cmd_reclassify)


def _add_common_args(parser: argparse.ArgumentParser) -> None:
    parser.add_argument(
        "--repo",
        default=os.environ.get("GITHUB_REPOSITORY", DEFAULT_REPO),
        help="owner/repo to query (default: %(default)s).",
    )
    parser.add_argument(
        "--data-dir",
        default=None,
        help=f"Metrics data directory (default: <repo>/{DATA_ROOT}).",
    )
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose logging.")


def _setup_logging(verbose: bool) -> None:
    logging.basicConfig(
        level=logging.DEBUG if verbose else logging.INFO,
        format="%(levelname)s %(message)s",
    )


def _store(args: argparse.Namespace) -> MetricsStore:
    root = Path(args.data_dir) if args.data_dir else REPO_ROOT / DATA_ROOT
    return MetricsStore(root=root, repo_root=REPO_ROOT)


def _parse_since(value: str) -> datetime:
    """Turn '90d' / '12w' / '3m' into an absolute UTC cutoff."""
    match = SINCE_RE.match(value.strip().lower())
    if not match:
        raise SystemExit(
            f"Invalid --since value {value!r} -- expected forms like 90d, 12w, 3m"
        )
    amount, unit = int(match.group(1)), match.group(2)
    days = {"d": 1, "w": 7, "m": 30}[unit] * amount
    return datetime.now(timezone.utc) - timedelta(days=days)


def _resolve_workflow(client: GitHubClient, name: str) -> Dict[str, Any]:
    target = WORKFLOW_ALIASES.get(name.lower(), name)
    workflows = client.list_workflows()
    for workflow in workflows:
        path = workflow.get("path", "")
        if path.endswith(target) or workflow.get("name", "").lower() == target.lower():
            return workflow
    known = ", ".join(sorted(WORKFLOW_ALIASES))
    raise SystemExit(f"Workflow {name!r} not found. Known aliases: {known}")


def _select_runs(
    client: GitHubClient, args: argparse.Namespace
) -> List[Dict[str, Any]]:
    if args.run_id:
        return [client.get_run(run_id) for run_id in args.run_id]

    if not args.workflow:
        raise SystemExit("Provide either --run-id or --workflow")

    workflow = _resolve_workflow(client, args.workflow)
    created = None
    if args.since:
        cutoff = _parse_since(args.since)
        created = f">={cutoff.date().isoformat()}"

    LOG.info(
        "Listing runs for %s%s",
        workflow.get("path"),
        f" since {created}" if created else "",
    )
    return client.list_workflow_runs(workflow["id"], created=created, limit=args.limit)


def cmd_ingest(args: argparse.Namespace) -> int:
    """Ingest one or more workflow runs (Tier 0: metadata only)."""
    _setup_logging(args.verbose)
    client = GitHubClient(repo=args.repo)
    store = _store(args)

    runs = _select_runs(client, args)
    if not runs:
        LOG.warning("No runs matched")
        return 0

    LOG.info("Ingesting %s run(s)", len(runs))
    ingested = skipped = 0

    for run in runs:
        run_id = run["id"]
        attempt = run.get("run_attempt", 1)
        slug = workflow_slug_for(run)
        month = month_for(run)

        if not args.force and store.has_run(slug, month, run_id, attempt):
            skipped += 1
            LOG.debug("Skipping already-ingested run %s attempt %s", run_id, attempt)
            continue

        try:
            record = ingest_run(
                client, run, attempt=attempt, detect_flakes=not args.no_flake_detection
            )
        except Exception as exc:  # keep a long backfill going past one bad run
            LOG.error("Failed to ingest run %s: %s", run_id, exc)
            continue

        store.write_run(
            workflow_slug(record),
            record_month(record),
            run_id,
            attempt,
            record.to_dict(),
        )
        ingested += 1
        LOG.info(
            "run %s attempt %s: %s jobs, %s failed, %s lost upstream",
            run_id,
            attempt,
            record.jobs_total,
            record.jobs_failure,
            record.jobs_not_run_due_to_upstream,
        )

    LOG.info(
        "Done: %s ingested, %s skipped, %s API calls",
        ingested,
        skipped,
        client.calls_made,
    )
    return 0


def workflow_slug_for(run: Dict[str, Any]) -> str:
    """Slug for a raw run payload (used before a record is built)."""
    source = (run.get("path") or run.get("name") or "unknown").split("@")[0]
    source = source.rsplit("/", 1)[-1]
    source = re.sub(r"\.ya?ml$", "", source)
    return re.sub(r"[^a-zA-Z0-9_.-]+", "-", source).strip("-").lower() or "unknown"


def month_for(run: Dict[str, Any]) -> str:
    created = run.get("created_at")
    if not created:
        return "unknown"
    try:
        return datetime.fromisoformat(created.replace("Z", "+00:00")).strftime("%Y-%m")
    except ValueError:
        return "unknown"


def _find_stored_run(
    store: MetricsStore, run_id: int, attempt: Optional[int]
) -> Optional[Dict[str, Any]]:
    suffix = f"{run_id}-{attempt}.json.gz" if attempt else f"{run_id}-"
    for path in store.iter_runs():
        if (attempt and path.name == suffix) or (
            not attempt and path.name.startswith(suffix)
        ):
            return store.read_run_path(path)
    return None


def cmd_show(args: argparse.Namespace) -> int:
    """Print a summary of a stored run record."""
    _setup_logging(args.verbose)
    store = _store(args)
    record = _find_stored_run(store, args.run_id, args.attempt)

    if record is None:
        LOG.error("No stored record for run %s (ingest it first)", args.run_id)
        return 1

    if args.json:
        json.dump(record, sys.stdout, indent=2, sort_keys=True)
        sys.stdout.write("\n")
        return 0

    print(
        f"{record['workflow_name']} run {record['run_id']} attempt {record['attempt']}"
    )
    print(f"  created    {record['created_at']}  conclusion={record['conclusion']}")
    print(
        "  jobs       {total} total / {ok} success / {fail} failure / "
        "{skip} skipped / {cancel} cancelled".format(
            total=record["jobs_total"],
            ok=record["jobs_success"],
            fail=record["jobs_failure"],
            skip=record["jobs_skipped"],
            cancel=record["jobs_cancelled"],
        )
    )
    if record["failed_prepare_jobs"]:
        print(
            f"  BLIND SPOT {len(record['failed_prepare_jobs'])} prepare job(s) failed -- "
            f"~{record['jobs_not_run_due_to_upstream']} test jobs never ran"
        )
    print(
        "  minutes    {total:.0f} total / {failed:.0f} wasted "
        "({sh:.0f} self-hosted wasted)".format(
            total=record["runner_minutes_total"],
            failed=record["runner_minutes_failed"],
            sh=record["self_hosted_minutes_failed"],
        )
    )
    print(f"  taxonomy   v{record.get('taxonomy_version', TAXONOMY_VERSION)}")

    failures = record.get("failures", [])
    if failures:
        print(f"\n  Failed jobs ({len(failures)}):")
        by_step: Dict[str, int] = {}
        for failure in failures:
            key = failure.get("failed_step_name") or "<no failed step>"
            by_step[key] = by_step.get(key, 0) + 1
        for step, count in sorted(by_step.items(), key=lambda kv: -kv[1]):
            print(f"    {count:4d}  {step}")
    return 0


def _iter_target_records(store: MetricsStore, args: argparse.Namespace) -> List[Path]:
    """Resolve the set of stored run records a command should operate on."""
    if getattr(args, "run_id", None):
        wanted = {str(run_id) for run_id in args.run_id}
        return [p for p in store.iter_runs() if p.name.split("-")[0] in wanted]
    if getattr(args, "workflow", None):
        return store.iter_runs(WORKFLOW_ALIASES.get(args.workflow, args.workflow))
    return store.iter_runs()


def cmd_classify(args: argparse.Namespace) -> int:
    """Apply the Stage A metadata router to stored failures.

    Classification is a separate pass over stored records rather than part of
    ingest. That separation is what makes the baseline honest: when the router
    or rule pack changes, history is recomputed from the same stored evidence
    instead of being refetched (and silently re-sampled) from GitHub.
    """
    _setup_logging(args.verbose)
    store = _store(args)

    paths = _iter_target_records(store, args)
    if not paths:
        LOG.warning("No stored runs matched -- ingest first")
        return 1

    pack = load_rules(Path(args.rules) if args.rules else None)
    totals: Dict[str, int] = {}
    by_rule: Dict[str, int] = {}
    total_failures = deferred = unresolved = 0

    for path in paths:
        record = store.read_run_path(path)
        if record is None:
            continue
        jobs = [_job_from_dict(f) for f in record.get("failures", [])]

        if args.job_id:
            jobs = [j for j in jobs if j.job_id == args.job_id]
            if not jobs:
                continue
            return _explain(jobs[0])

        summary = route_all(jobs)
        total_failures += summary["total"]
        for key, count in summary["by_class"].items():
            totals[key] = totals.get(key, 0) + count

        # Stage B settles what Stage A deferred, using the stored excerpt.
        rule_summary = classify_with_rules(jobs, pack)
        for job in jobs:
            if job.classified_by == ClassifiedBy.RULES.value:
                key = f"{job.failure_class}/{job.subclass}"
                totals[key] = totals.get(key, 0) + 1
                deferred_class = f"{FailureClass.UNKNOWN.value}/None"
                if deferred_class in totals:
                    totals[deferred_class] -= 1
        for rule_id, count in rule_summary["by_rule"].items():
            by_rule[rule_id] = by_rule.get(rule_id, 0) + count
        unresolved += rule_summary["unmatched"]
        deferred += summary["deferred_to_stage_b"]

        if not args.dry_run:
            record["failures"] = [j.to_dict() for j in jobs]
            record["router_version"] = ROUTER_VERSION
            record["ruleset_version"] = pack.version
            store.write_run_path(path, record)

    if args.job_id:
        LOG.error("Job %s not found in the stored records", args.job_id)
        return 1

    print(f"Classified {total_failures} failed job(s) across {len(paths)} run(s)")
    for key, count in sorted(totals.items(), key=lambda kv: -kv[1]):
        share = _percent(count, total_failures)
        print(f"  {count:5d}  {share:>6}  {key}")
    print(
        f"\nStage A deferred {deferred} -- Stage B (rules v{pack.version}) "
        f"resolved {deferred - unresolved}, left {unresolved} unclassified "
        f"({_percent(unresolved, total_failures)} of all failures)"
    )
    if by_rule:
        print("\nBy rule:")
        for rule_id, count in sorted(by_rule.items(), key=lambda kv: -kv[1]):
            print(f"  {count:5d}  {rule_id}")
    if args.dry_run:
        print("\n(dry run -- nothing written)")
    return 0


def _percent(count: int, total: int) -> str:
    return f"{(100.0 * count / total) if total else 0.0:.1f}%"


def _explain(job: JobFailure) -> int:
    """Print a single job's routing verdict and the evidence behind it."""
    verdict = route(job)
    print(f"job {job.job_id}  {job.job_name}")
    print(f"  url          {job.html_url}")
    print(f"  failed step  {job.failed_step_name or '<none>'}")
    print(f"  steps        {len(job.steps)} recorded")
    null_steps = [s.name for s in job.steps if s.conclusion is None]
    if null_steps:
        print(f"  never ran    {len(null_steps)} step(s), first: {null_steps[0]}")
    print(f"  log blob     {'present' if job.log_available else 'absent/unknown'}")
    print(f"  duration     {job.duration_s}s")
    print()
    if verdict.defer:
        print(f"  VERDICT      deferred to Stage B (rule {verdict.rule_id})")
    else:
        print(f"  VERDICT      {verdict.failure_class}/{verdict.subclass}")
        print(f"  rule         {verdict.rule_id}  confidence {verdict.confidence}")
    return 0


def _job_from_dict(payload: Dict[str, Any]) -> JobFailure:
    """Rebuild a JobFailure from its stored dict form."""
    steps = [StepRecord(**s) for s in payload.get("steps", [])]
    fields = {k: v for k, v in payload.items() if k in JobFailure.__annotations__}
    fields["steps"] = steps
    return JobFailure(**fields)


def cmd_fetch_logs(args: argparse.Namespace) -> int:
    """Tier 1: fetch logs for failures the router deferred, and fingerprint them.

    Only deferred failures are fetched. Failures the router already resolved
    from metadata need no log, which is what keeps this affordable: a nightly
    run has ~2,500 jobs, ~75 failures, and only the deferred subset -- around
    70 -- costs a download.

    A 404 is recorded rather than retried. GitHub serves no log blob for a job
    whose runner disappeared, so the absence *is* the evidence; it upgrades the
    router's runner_lost verdict from inferred to confirmed.
    """
    _setup_logging(args.verbose)
    client = GitHubClient(repo=args.repo)
    store = _store(args)

    paths = _iter_target_records(store, args)
    if not paths:
        LOG.warning("No stored runs matched -- ingest first")
        return 1

    fetched = skipped = missing = 0
    signatures: Dict[str, int] = {}

    for path in paths:
        record = store.read_run_path(path)
        if record is None:
            continue
        jobs = [_job_from_dict(f) for f in record.get("failures", [])]

        targets = []
        for job in jobs:
            verdict = apply_router(job)
            if not verdict.defer:
                continue
            if job.signature_id and not args.force:
                skipped += 1
                continue
            targets.append(job)

        if targets:
            LOG.info(
                "run %s: fetching %s log(s) of %s failure(s)",
                record["run_id"],
                len(targets),
                len(jobs),
            )
            _fetch_into(client, targets, args.workers)

        for job in jobs:
            if job.signature_id:
                signatures[job.signature_id] = signatures.get(job.signature_id, 0) + 1
                if job.log_available:
                    fetched += 1
            elif job.log_available is False:
                missing += 1

        record["failures"] = [j.to_dict() for j in jobs]
        record["normaliser_version"] = NORMALISER_VERSION
        store.write_run_path(path, record)

    print(f"Fingerprinted {fetched} failure(s) across {len(paths)} run(s)")
    print(f"  {skipped} already had a signature, {missing} had no log blob")
    print(f"  {len(signatures)} distinct signature(s)")
    for sig, count in sorted(signatures.items(), key=lambda kv: -kv[1])[:15]:
        print(f"    {count:5d}  {sig}")
    return 0


def _fetch_into(client: GitHubClient, jobs: List[JobFailure], workers: int) -> None:
    """Download and fingerprint logs concurrently, writing onto each job."""

    def work(job: JobFailure) -> None:
        log = client.get_job_log(job.job_id)
        if log is None:
            # No blob: the runner vanished. Recording this lets the router
            # raise its confidence on the next pass instead of guessing.
            job.log_available = False
            return
        job.log_available = True
        fp = fingerprint(log)
        job.signature_id = fp.signature_id
        job.evidence_excerpt = fp.excerpt
        job.normaliser_version = fp.normaliser_version

    with ThreadPoolExecutor(max_workers=max(1, workers)) as pool:
        for future in as_completed([pool.submit(work, job) for job in jobs]):
            try:
                future.result()
            except Exception as exc:  # one bad log must not sink the run
                LOG.warning("Log fetch failed: %s", exc)


def cmd_audit_secrets(args: argparse.Namespace) -> int:
    """Scan every stored excerpt for secrets.

    This is the hard gate from the rollout plan. It runs over what was
    *actually persisted*, independently of the scrubber having been invoked,
    so a scrubber regression cannot hide behind its own output.
    """
    _setup_logging(args.verbose)
    store = _store(args)

    scanned = 0
    findings: List[str] = []

    for path in store.iter_runs():
        record = store.read_run_path(path)
        if record is None:
            continue
        for failure in record.get("failures", []):
            excerpt = failure.get("evidence_excerpt")
            if not excerpt:
                continue
            scanned += 1
            for line_no, rule in find_secrets(excerpt):
                findings.append(
                    f"{path.name} job {failure['job_id']} line {line_no}: {rule}"
                )

    for path in sorted(store.root.rglob("*.json")):
        text = path.read_text(encoding="utf-8", errors="replace")
        for line_no, rule in find_secrets(text):
            findings.append(f"{path.name} line {line_no}: {rule}")
        scanned += 1

    print(f"Scanned {scanned} stored excerpt(s)/file(s)")
    if findings:
        print(f"\nFAIL: {len(findings)} finding(s)")
        for finding in findings[:50]:
            print(f"  {finding}")
        return 1
    print("PASS: no secrets found")
    return 0


if __name__ == "__main__":
    _parser = argparse.ArgumentParser()
    _sub = _parser.add_subparsers(dest="subcommand", required=True)
    add_metrics_cmds(_sub)
    _args = _parser.parse_args()
    sys.exit(_args.func(_args))


def cmd_rollup(args: argparse.Namespace) -> int:
    """Aggregate stored run records into rollups, one per period.

    Rollups are computed from stored records only. Nothing here touches the
    API, so a rollup can always be regenerated and audited after the fact.
    """
    _setup_logging(args.verbose)
    store = _store(args)

    slug = None
    if args.workflow:
        slug = WORKFLOW_ALIASES.get(args.workflow, args.workflow)
        slug = re.sub(r"\.ya?ml$", "", slug.rsplit("/", 1)[-1])

    by_period: Dict[str, List[Dict[str, Any]]] = {}
    for path in store.iter_runs(slug):
        record = store.read_run_path(path)
        if record is None:
            continue
        # The stored layout is runs/<workflow>/<YYYY-MM>/<run>-<attempt>.json.gz,
        # so the period is the parent directory rather than a re-parsed date.
        period = path.parent.name
        record.setdefault("workflow_slug", path.parent.parent.name)
        if args.period and period != args.period:
            continue
        by_period.setdefault(period, []).append(record)

    if not by_period:
        LOG.warning("No stored runs matched -- ingest and classify first")
        return 1

    pack = load_rules()
    for period, records in sorted(by_period.items()):
        payload = aggregate(records, period)
        payload["ruleset_version"] = pack.version
        if args.dry_run:
            print(json.dumps(payload, indent=2, sort_keys=True))
            continue
        path = store.write_rollup(period, payload)
        LOG.info(
            "%s: %s run(s), %s failure(s), %s signature(s) -> %s",
            period,
            payload["runs"],
            payload["jobs_failure"],
            len(payload["signatures"]),
            path,
        )
    return 0


def cmd_report(args: argparse.Namespace) -> int:
    """Render a digest or trend report from stored rollups."""
    _setup_logging(args.verbose)
    store = _store(args)

    paths = sorted((store.root / "rollups").glob("*.json"))
    if not paths:
        LOG.error("No rollups found -- run `metrics rollup` first")
        return 1

    rollups = [store.read_json(p) for p in paths]
    rollups = [r for r in rollups if r]

    if args.format == "markdown":
        start = max(0, len(rollups) - args.periods)
        text = render_markdown(rollups[start:])
    else:
        if args.period:
            selected = [r for r in rollups if r.get("period") == args.period]
            if not selected:
                LOG.error("No rollup for period %s", args.period)
                return 1
            index = rollups.index(selected[0])
        else:
            index = len(rollups) - 1
        previous = rollups[index - 1] if index > 0 else None
        text = render_mattermost(rollups[index], previous)

    if args.out:
        Path(args.out).write_text(text)
        LOG.info("Wrote %s", args.out)
    else:
        print(text)
    return 0


def cmd_reclassify(args: argparse.Namespace) -> int:
    """Replay the rule pack over every stored failure.

    This is what makes the rule pack safe to iterate on: rules can be added or
    corrected and the entire retained history re-derived from stored excerpts,
    with no API traffic and no dependence on job logs that GitHub has since
    expired.

    Both stages are replayed, in order. Stage A reads job metadata, which is
    stored in full, so it is as replayable as Stage B -- and running it here
    means a record that was ingested but never classified gets its router
    verdict without any API traffic. Stage B then settles whatever Stage A
    deferred.
    """
    _setup_logging(args.verbose)
    store = _store(args)
    pack = load_rules(Path(args.rules) if args.rules else None)
    if not pack.rules:
        LOG.error("Rule pack is empty -- nothing to replay")
        return 1

    slug = None
    if args.workflow:
        slug = WORKFLOW_ALIASES.get(args.workflow, args.workflow)
        slug = re.sub(r"\.ya?ml$", "", slug.rsplit("/", 1)[-1])

    changed_records = 0
    changed = collections.Counter()
    unchanged = 0
    for path in store.iter_runs(slug):
        if args.period and path.parent.name != args.period:
            continue
        record = store.read_run_path(path)
        if record is None:
            continue

        dirty = False
        for failure in record.get("failures") or []:
            job = JobFailure.from_dict(failure)
            before = (job.failure_class, job.subclass, job.rule_id)
            # Reset before replaying. Without this, a rule deleted from the
            # pack would leave its stale attribution behind forever, and
            # reclassification would only ever be additive.
            job.failure_class = FailureClass.UNKNOWN.value
            job.subclass = None
            job.rule_id = None
            job.classified_by = ClassifiedBy.NONE.value
            apply_router(job)
            if job.classified_by != ClassifiedBy.ROUTER.value:
                pack.apply(job)
            after = (job.failure_class, job.subclass, job.rule_id)
            if before == after:
                unchanged += 1
                continue
            changed[f"{before[0]} -> {after[0]}"] += 1
            failure.update(job.to_dict())
            dirty = True

        if dirty:
            changed_records += 1
            if not args.dry_run:
                store.write_run_path(path, record)

    verb = "would change" if args.dry_run else "changed"
    LOG.info(
        "%s %s failure(s) across %s run record(s); %s unchanged",
        verb,
        sum(changed.values()),
        changed_records,
        unchanged,
    )
    for transition, count in changed.most_common():
        LOG.info("  %6d  %s", count, transition)
    return 0
