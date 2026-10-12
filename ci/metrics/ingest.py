#
# Copyright 2026 Canonical, Ltd.
#
"""Ingest GitHub Actions runs and jobs into the metrics schema."""

from __future__ import annotations

import dataclasses
import re
import sqlite3
import statistics
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

from . import INGEST_VERSION
from . import store as metrics_store
from .gh import GitHubClient

InFlightRunError = metrics_store.InFlightRunError

NAME_SPLIT = " / "
ARCHES = {"amd64", "arm64", "ppc64el", "s390x"}
UBUNTU_SERIES = {"focal", "jammy", "noble", "oracular", "plucky", "questing"}
LXD_IMAGE_RE = re.compile(r"^[a-z]+:\d+\.\d+$")
CHANNEL_RE = re.compile(
    r"^(latest|\d+\.\d+(?:-classic|-strict)?)/(stable|candidate|beta|edge)$"
)
FLAVOR_RE = re.compile(r"^(classic|strict)$")
PREPARE_JOB_NAME = "Prepare Environment"
TEST_NODEID_RE = re.compile(r"^(?P<file>[\w/.\-]+\.py)(?:::(?P<test>.+))?$")


@dataclass(frozen=True)
class StepRecord:
    name: str
    number: int
    conclusion: Optional[str]
    duration_s: Optional[float] = None


@dataclass
class ParsedJobName:
    raw: str
    caller_job: Optional[str] = None
    inner_name: Optional[str] = None
    os: str = ""
    arch: str = ""
    channel: str = ""
    flavor: str = ""
    substrate: str = ""
    test_nodeid: str = ""
    test_file: Optional[str] = None
    unparsed_params: List[str] = field(default_factory=list)

    @property
    def caller_base(self) -> Optional[str]:
        if not self.caller_job:
            return None
        return re.sub(r"\s*\(.*\)\s*$", "", self.caller_job).strip()

    @property
    def config_key(self) -> Tuple[str, str, str, str, str]:
        return (self.os, self.arch, self.channel, self.flavor, self.substrate)

    @property
    def is_prepare(self) -> bool:
        return self.inner_name == PREPARE_JOB_NAME


@dataclass(frozen=True)
class JobRecord:
    job_id: int
    run_id: int
    attempt: int
    config_key: Tuple[str, str, str, str, str]
    test_nodeid: str
    test_file: Optional[str]
    runner_labels: Tuple[str, ...]
    self_hosted: bool
    job_name: str
    caller_job: Optional[str]
    html_url: str
    started_at: Optional[str]
    completed_at: Optional[str]
    duration_s: Optional[float]
    failed_step_name: Optional[str]
    failed_step_number: Optional[int]
    log_available: Optional[int]
    steps: Tuple[StepRecord, ...]
    tested_revision: str = ""
    tested_artifact: str = ""
    collection_ref: str = ""
    collection_sha: str = ""


@dataclass(frozen=True)
class RetryOutcomeRecord:
    run_id: int
    attempt: int
    config_key: Tuple[str, str, str, str, str]
    test_nodeid: str
    test_file: Optional[str]
    conclusion: str


@dataclass
class RunRecord:
    run_id: int
    attempt: int
    workflow_name: str
    workflow_file: Optional[str]
    event: Optional[str]
    head_branch: Optional[str]
    head_sha: Optional[str]
    pr_number: Optional[int]
    created_at: Optional[str]
    started_at: Optional[str]
    completed_at: str
    conclusion: str
    jobs_total: int
    jobs_success: int
    jobs_failure: int
    jobs_skipped: int
    jobs_cancelled: int
    runner_minutes_total: float
    runner_minutes_failed: float
    failed_jobs: List[JobRecord] = field(default_factory=list)
    tests_collected: Dict[Tuple[str, str, str, str, str], int] = field(
        default_factory=dict
    )
    blocked_jobs: Dict[str, int] = field(default_factory=dict)
    retry_outcomes: List[RetryOutcomeRecord] = field(default_factory=list)
    ingest_version: int = INGEST_VERSION


def _classify_param(value: str, parsed: ParsedJobName) -> None:
    value = value.strip()
    if not value:
        return
    if value in ARCHES:
        parsed.arch = value
    elif CHANNEL_RE.match(value):
        parsed.channel = value
    elif LXD_IMAGE_RE.match(value):
        parsed.os = value
        parsed.substrate = "lxd"
    elif value in UBUNTU_SERIES:
        parsed.os = value
        parsed.substrate = "multipass"
    elif FLAVOR_RE.match(value):
        parsed.flavor = value
    else:
        parsed.unparsed_params.append(value)


def parse_job_name(name: str) -> ParsedJobName:
    """Recover matrix dimensions and test identity from a GitHub job name."""

    parsed = ParsedJobName(raw=name)
    head, _, tail = name.partition(NAME_SPLIT)
    if tail:
        parsed.caller_job = head.strip()
        parsed.inner_name = tail.strip()
    else:
        parsed.inner_name = head.strip()

    source = parsed.caller_job or ""
    match = re.search(r"\((.*)\)\s*$", source)
    if match:
        for param in match.group(1).split(","):
            _classify_param(param, parsed)
    elif source:
        for token in re.split(r"[\s\-]+", source):
            if token in ARCHES:
                parsed.arch = token

    inner = parsed.inner_name or ""
    node_match = TEST_NODEID_RE.match(inner)
    if node_match and node_match.group("test"):
        parsed.test_nodeid = inner
        parsed.test_file = node_match.group("file")
    elif node_match:
        parsed.test_file = node_match.group("file")

    if not parsed.substrate and (parsed.test_nodeid or parsed.test_file):
        parsed.substrate = "lxd"

    return parsed


def _parse_ts(value: Optional[str]) -> Optional[datetime]:
    if not value:
        return None
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None


def _duration_s(started: Optional[str], completed: Optional[str]) -> Optional[float]:
    start, end = _parse_ts(started), _parse_ts(completed)
    if not start or not end:
        return None
    return max(0.0, (end - start).total_seconds())


def _step_records(job: Dict[str, Any]) -> List[StepRecord]:
    return [
        StepRecord(
            name=step.get("name", ""),
            number=step.get("number", 0),
            conclusion=step.get("conclusion"),
            duration_s=_duration_s(step.get("started_at"), step.get("completed_at")),
        )
        for step in job.get("steps") or []
    ]


def find_failed_step(steps: Iterable[StepRecord]) -> Optional[StepRecord]:
    for step in steps:
        if step.conclusion == "failure":
            return step
    return None


def _is_self_hosted(labels: Iterable[str]) -> bool:
    labels_set = set(labels)
    return "self-hosted" in labels_set or "self_hosted" in labels_set


def _label_key(labels: Iterable[str]) -> str:
    return ",".join(sorted(label for label in labels if label))


def _normalise_conclusion(value: Optional[str]) -> str:
    if value in {"success", "failure", "skipped", "cancelled"}:
        return value
    return "cancelled"


def _job_record(job: Dict[str, Any], parsed: ParsedJobName) -> JobRecord:
    steps = _step_records(job)
    failed_step = find_failed_step(steps)
    labels = tuple(job.get("labels") or ())
    arch = parsed.arch or next((label for label in labels if label in ARCHES), "")
    config_key = (
        parsed.os,
        arch,
        parsed.channel,
        parsed.flavor,
        parsed.substrate,
    )
    return JobRecord(
        job_id=job["id"],
        run_id=job["run_id"],
        attempt=job.get("run_attempt", 1),
        config_key=config_key,
        test_nodeid=parsed.test_nodeid,
        test_file=parsed.test_file,
        runner_labels=labels,
        self_hosted=_is_self_hosted(labels),
        job_name=job.get("name", ""),
        caller_job=parsed.caller_base,
        html_url=job.get("html_url", ""),
        started_at=job.get("started_at"),
        completed_at=job.get("completed_at"),
        duration_s=_duration_s(job.get("started_at"), job.get("completed_at")),
        failed_step_name=failed_step.name if failed_step else None,
        failed_step_number=failed_step.number if failed_step else None,
        log_available=None,
        steps=tuple(steps),
    )


def _summarise_jobs(
    jobs: List[Dict[str, Any]]
) -> Tuple[
    Dict[str, int], Dict[Tuple[str, str, str, str, str], int], List[ParsedJobName]
]:
    counts = {"success": 0, "failure": 0, "skipped": 0, "cancelled": 0}
    collected: Dict[Tuple[str, str, str, str, str], int] = {}
    failed_prepare: List[ParsedJobName] = []

    for job in jobs:
        conclusion = _normalise_conclusion(job.get("conclusion"))
        counts[conclusion] += 1
        parsed = parse_job_name(job.get("name", ""))
        if parsed.is_prepare:
            if conclusion == "failure":
                failed_prepare.append(parsed)
            continue
        if parsed.test_file:
            collected[parsed.config_key] = collected.get(parsed.config_key, 0) + 1

    return counts, collected, failed_prepare


def _estimate_lost_jobs(
    failed_prepare: List[ParsedJobName],
    collected: Dict[Tuple[str, str, str, str, str], int],
) -> Dict[str, int]:
    by_caller: Dict[str, List[int]] = {}
    for key, count in collected.items():
        if count:
            by_caller.setdefault("|".join(key), []).append(count)

    blocked: Dict[str, int] = {}
    all_counts = [count for count in collected.values() if count]
    fallback = int(statistics.median(all_counts)) if all_counts else 0
    for parsed in failed_prepare:
        siblings = by_caller.get("|".join(parsed.config_key), [])
        blocked[parsed.raw] = int(statistics.median(siblings)) if siblings else fallback
    return blocked


def _accumulate_minutes(jobs: List[Dict[str, Any]]) -> Dict[str, float]:
    totals = {"runner_minutes_total": 0.0, "runner_minutes_failed": 0.0}
    for job in jobs:
        duration = _duration_s(job.get("started_at"), job.get("completed_at")) or 0.0
        minutes = duration / 60.0
        totals["runner_minutes_total"] += minutes
        if job.get("conclusion") == "failure":
            totals["runner_minutes_failed"] += minutes
    return totals


def _pr_number(run: Dict[str, Any]) -> Optional[int]:
    pull_requests = run.get("pull_requests") or []
    if not pull_requests:
        return None
    return pull_requests[0].get("number")


def _workflow_file(run: Dict[str, Any]) -> Optional[str]:
    path = (run.get("path") or "").split("@")[0]
    return path or None


def _ensure_completed(run: Dict[str, Any]) -> None:
    if not run.get("conclusion"):
        raise InFlightRunError(
            f"Refusing to store in-flight run {run.get('id')}: conclusion is empty"
        )
    if not run.get("updated_at"):
        raise InFlightRunError(
            f"Refusing to store run {run.get('id')}: completed timestamp is empty"
        )


def build_run_record(
    client: GitHubClient, run: Dict[str, Any], attempt: int
) -> RunRecord:
    """Build one immutable run-attempt record from GitHub job metadata."""

    _ensure_completed(run)
    run_id = run["id"]
    jobs = client.list_run_jobs(run_id, attempt=attempt)
    counts, collected, failed_prepare = _summarise_jobs(jobs)
    minutes = _accumulate_minutes(jobs)

    failures: List[JobRecord] = []
    for job in jobs:
        if job.get("conclusion") != "failure":
            continue
        parsed = parse_job_name(job.get("name", ""))
        failures.append(_job_record(job, parsed))

    return RunRecord(
        run_id=run_id,
        attempt=attempt,
        workflow_name=run.get("name", ""),
        workflow_file=_workflow_file(run),
        event=run.get("event"),
        head_branch=run.get("head_branch"),
        head_sha=run.get("head_sha"),
        pr_number=_pr_number(run),
        created_at=run.get("created_at"),
        started_at=run.get("run_started_at") or run.get("created_at"),
        completed_at=run["updated_at"],
        conclusion=run["conclusion"],
        jobs_total=len(jobs),
        jobs_success=counts["success"],
        jobs_failure=counts["failure"],
        jobs_skipped=counts["skipped"],
        jobs_cancelled=counts["cancelled"],
        runner_minutes_total=round(minutes["runner_minutes_total"], 2),
        runner_minutes_failed=round(minutes["runner_minutes_failed"], 2),
        failed_jobs=failures,
        tests_collected=collected,
        blocked_jobs=_estimate_lost_jobs(failed_prepare, collected),
    )


def build_run_records(client: GitHubClient, run: Dict[str, Any]) -> List[RunRecord]:
    """Build records for every attempt of a completed workflow run."""

    _ensure_completed(run)
    latest_attempt = int(run.get("run_attempt") or 1)
    records = [
        build_run_record(client, run, attempt=attempt)
        for attempt in range(1, latest_attempt + 1)
    ]
    _attach_retry_outcomes(client, records)
    return records


def _attach_retry_outcomes(client: GitHubClient, records: List[RunRecord]) -> None:
    failures_by_attempt = {
        record.attempt: {
            (job.config_key, job.test_nodeid, job.test_file)
            for job in record.failed_jobs
        }
        for record in records
    }
    for record in records:
        previous_failures = set()
        for previous in records:
            if previous.attempt < record.attempt:
                previous_failures.update(failures_by_attempt[previous.attempt])
        if not previous_failures:
            continue

        # Use the full retry attempt job list, not absence from failed jobs.
        jobs = client.list_run_jobs(record.run_id, attempt=record.attempt)
        for job in jobs:
            parsed = parse_job_name(job.get("name", ""))
            key = (parsed.config_key, parsed.test_nodeid, parsed.test_file)
            if key not in previous_failures:
                continue
            record.retry_outcomes.append(
                RetryOutcomeRecord(
                    run_id=record.run_id,
                    attempt=record.attempt,
                    config_key=parsed.config_key,
                    test_nodeid=parsed.test_nodeid,
                    test_file=parsed.test_file,
                    conclusion=_normalise_conclusion(job.get("conclusion")),
                )
            )


def initialise_database(conn: sqlite3.Connection) -> None:
    metrics_store.initialise_database(conn)


def open_database(path: Path) -> sqlite3.Connection:
    return metrics_store.open_database(path)


def store_run_record(conn: sqlite3.Connection, record: RunRecord) -> None:
    metrics_store.store_run_record(conn, record)


def ingest_run(
    client: GitHubClient, conn: sqlite3.Connection, run_id: int
) -> List[RunRecord]:
    """Fetch and store every attempt of one completed workflow run."""

    run = client.get_run(run_id)
    records = build_run_records(client, run)
    with conn:
        for record in records:
            store_run_record(conn, record)
    return records


def dump_records(records: Iterable[RunRecord]) -> List[Dict[str, Any]]:
    """Return a compact serialisable summary for CLI output and tests."""

    return [
        {
            **dataclasses.asdict(record),
            "failed_jobs": len(record.failed_jobs),
            "tests_collected": len(record.tests_collected),
            "blocked_jobs": len(record.blocked_jobs),
            "retry_outcomes": len(record.retry_outcomes),
        }
        for record in records
    ]
