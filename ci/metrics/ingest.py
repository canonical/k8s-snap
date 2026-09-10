#
# Copyright 2026 Canonical, Ltd.
#
"""Tier-0 ingest: GitHub Actions job metadata into :class:`RunRecord`.

This tier never reads a log. Job and step metadata alone answer the single most
valuable question -- infrastructure versus product -- and cost 26 API calls for
a 2555-job nightly run.

Two integrity concerns are handled here rather than left to the reporting layer:

* **Denominator drift.** Each channel collects its tests from the matching
  release branch, so the number of tests per (os, arch, channel) varies
  (42-68 observed in one run). Raw failure counts are therefore not comparable
  across runs; ``tests_collected_by_combo`` records the denominator.
* **The upstream blind spot.** When a ``Prepare Environment`` job fails, its
  entire downstream test matrix never runs and so never appears as failed.
  Left unmeasured, a prepare failure makes the metrics look *better*.
  ``jobs_not_run_due_to_upstream`` counts the disappeared jobs.
"""

import logging
import re
import statistics
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Optional, Tuple

from . import INGEST_VERSION, TAXONOMY_VERSION
from .gh import GitHubClient
from .models import JobFailure, RunRecord, StepRecord

LOG = logging.getLogger(__name__)

# GitHub renders reusable-workflow jobs as "<caller job> / <inner job name>".
NAME_SPLIT = " / "

ARCHES = {"amd64", "arm64", "ppc64el", "s390x"}

# Ubuntu series names appear where a Multipass image is used; the LXD path
# uses "ubuntu:24.04" style images instead. The distinction is a reliable
# substrate signal, so substrate is inferred rather than guessed from workflow.
UBUNTU_SERIES = {"focal", "jammy", "noble", "oracular", "plucky", "questing"}

LXD_IMAGE_RE = re.compile(r"^[a-z]+:\d+\.\d+$")
CHANNEL_RE = re.compile(
    r"^(latest|\d+\.\d+(?:-classic|-strict)?)/(stable|candidate|beta|edge)$"
)
FLAVOR_RE = re.compile(r"^(classic|strict)$")

PREPARE_JOB_NAME = "Prepare Environment"

# A pytest node id, optionally parametrised, e.g.
# "tests/test_node_availability_zone.py::test_node_availability_zone[etcd-True]".
TEST_NODEID_RE = re.compile(r"^(?P<file>[\w/.\-]+\.py)(?:::(?P<test>.+))?$")


@dataclass
class ParsedJobName:
    """Dimensions recovered from a job name.

    Matrix parameters are identified by *content*, not position: the number and
    order of parameters differs between callers (``Integration (amd64)`` on
    PRs, ``Integration (os, arch, channel)`` nightly, and
    ``Spread tests across different datastores (amd64, latest/edge)`` weekly,
    which omits the OS entirely). Positional parsing would silently mis-assign
    these.
    """

    raw: str
    caller_job: Optional[str] = None
    inner_name: Optional[str] = None
    os: Optional[str] = None
    arch: Optional[str] = None
    channel: Optional[str] = None
    flavor: Optional[str] = None
    substrate: Optional[str] = None
    test_nodeid: Optional[str] = None
    test_file: Optional[str] = None
    unparsed_params: List[str] = field(default_factory=list)

    @property
    def combo(self) -> str:
        """Stable key for the matrix cell this job belongs to."""
        return "|".join(
            [
                self.caller_base or "",
                self.os or "",
                self.arch or "",
                self.channel or "",
            ]
        )

    @property
    def caller_base(self) -> Optional[str]:
        """Caller job name with its matrix parameters stripped."""
        if not self.caller_job:
            return None
        return re.sub(r"\s*\(.*\)\s*$", "", self.caller_job).strip()

    @property
    def is_prepare(self) -> bool:
        return self.inner_name == PREPARE_JOB_NAME


def _classify_param(value: str, parsed: ParsedJobName) -> None:
    """Assign one matrix parameter to a dimension based on its shape."""
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
    """Recover os/arch/channel/test identity from a GitHub job name.

    Handles every shape observed in this repository, including non-matrix jobs
    ("Get e2e test tags"), jobs whose matrix produced an empty parameter
    ("Build k8s-snap  amd64"), and parametrised pytest ids.
    """
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
        # Callers such as "Build k8s-snap  amd64" or "Charm e2e tests - amd64"
        # interpolate matrix values without parentheses.
        for token in re.split(r"[\s\-]+", source):
            if token in ARCHES:
                parsed.arch = token

    inner = parsed.inner_name or ""
    node_match = TEST_NODEID_RE.match(inner)
    if node_match and node_match.group("test"):
        parsed.test_nodeid = inner
        parsed.test_file = node_match.group("file")
    elif node_match:
        # PR runs schedule a whole file per job rather than a single test.
        parsed.test_file = node_match.group("file")

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
    records = []
    for step in job.get("steps") or []:
        records.append(
            StepRecord(
                name=step.get("name", ""),
                number=step.get("number", 0),
                conclusion=step.get("conclusion"),
                duration_s=_duration_s(
                    step.get("started_at"), step.get("completed_at")
                ),
            )
        )
    return records


def find_failed_step(steps: List[StepRecord]) -> Optional[StepRecord]:
    """Return the first step that failed, if any.

    A failed job with *no* failed step is not an anomaly to paper over: it is
    the signature of a runner that vanished mid-execution, and it is detected
    downstream by the Stage A router.
    """
    for step in steps:
        if step.conclusion == "failure":
            return step
    return None


def _is_self_hosted(labels: List[str]) -> bool:
    return any(label == "self-hosted" for label in labels)


def build_job_failure(job: Dict[str, Any], parsed: ParsedJobName) -> JobFailure:
    steps = _step_records(job)
    failed_step = find_failed_step(steps)
    labels = list(job.get("labels") or [])

    return JobFailure(
        job_id=job["id"],
        run_id=job["run_id"],
        attempt=job.get("run_attempt", 1),
        job_name=job.get("name", ""),
        html_url=job.get("html_url", ""),
        caller_job=parsed.caller_base,
        os=parsed.os,
        arch=parsed.arch or (labels[-1] if labels and labels[-1] in ARCHES else None),
        channel=parsed.channel,
        flavor=parsed.flavor,
        substrate=parsed.substrate,
        test_nodeid=parsed.test_nodeid,
        test_file=parsed.test_file,
        runner_labels=labels,
        self_hosted=_is_self_hosted(labels),
        started_at=job.get("started_at"),
        completed_at=job.get("completed_at"),
        duration_s=_duration_s(job.get("started_at"), job.get("completed_at")),
        failed_step_name=failed_step.name if failed_step else None,
        failed_step_number=failed_step.number if failed_step else None,
        steps=steps,
    )


def _estimate_lost_jobs(
    failed_prepare: List[ParsedJobName], collected: Dict[str, int]
) -> int:
    """Estimate how many test jobs never ran because their prepare job failed.

    There is no way to know the exact number: the test list is produced *by*
    the job that failed. The median sibling matrix cell of the same caller is
    the most defensible estimate available, and reporting it as an estimate is
    far better than silently recording zero.
    """
    if not failed_prepare:
        return 0

    by_caller: Dict[str, List[int]] = {}
    for combo, count in collected.items():
        caller = combo.split("|")[0]
        if count:
            by_caller.setdefault(caller, []).append(count)

    total = 0
    for parsed in failed_prepare:
        siblings = by_caller.get(parsed.caller_base or "", [])
        if siblings:
            total += int(statistics.median(siblings))
    return total


def summarise_jobs(
    jobs: List[Dict[str, Any]]
) -> Tuple[Dict[str, int], Dict[str, int], List[ParsedJobName]]:
    """Return (conclusion counts, tests-collected per combo, failed prepare jobs)."""
    counts = {"success": 0, "failure": 0, "skipped": 0, "cancelled": 0, "other": 0}
    collected: Dict[str, int] = {}
    failed_prepare: List[ParsedJobName] = []

    for job in jobs:
        conclusion = job.get("conclusion") or "other"
        counts[conclusion if conclusion in counts else "other"] += 1

        parsed = parse_job_name(job.get("name", ""))
        if parsed.is_prepare:
            if conclusion == "failure":
                failed_prepare.append(parsed)
            continue
        if parsed.test_file:
            collected[parsed.combo] = collected.get(parsed.combo, 0) + 1

    return counts, collected, failed_prepare


def _accumulate_minutes(jobs: List[Dict[str, Any]]) -> Dict[str, float]:
    totals = {
        "runner_minutes_total": 0.0,
        "runner_minutes_failed": 0.0,
        "self_hosted_minutes": 0.0,
        "self_hosted_minutes_failed": 0.0,
    }
    for job in jobs:
        duration = _duration_s(job.get("started_at"), job.get("completed_at")) or 0.0
        minutes = duration / 60.0
        self_hosted = _is_self_hosted(list(job.get("labels") or []))
        failed = job.get("conclusion") == "failure"

        totals["runner_minutes_total"] += minutes
        if failed:
            totals["runner_minutes_failed"] += minutes
        if self_hosted:
            totals["self_hosted_minutes"] += minutes
            if failed:
                totals["self_hosted_minutes_failed"] += minutes
    return totals


def annotate_retries(
    client: GitHubClient, run_id: int, attempt: int, failures: List[JobFailure]
) -> None:
    """Mark failures that passed on a later attempt of the same commit.

    This is ground-truth flake detection: same SHA, same job, different
    outcome. Verified against run 34237151320, where attempt 1 failed
    test_config_propagation and attempt 2 passed with no code change.
    """
    if attempt < 1 or not failures:
        return

    later_outcomes: Dict[str, str] = {}
    for later in range(attempt + 1, attempt + 6):
        try:
            jobs = client.list_run_jobs(run_id, attempt=later)
        except Exception as exc:  # pragma: no cover - defensive
            LOG.debug("No attempt %s for run %s (%s)", later, run_id, exc)
            break
        if not jobs:
            break
        for job in jobs:
            name = job.get("name", "")
            # Keep the earliest later-attempt outcome for each job.
            later_outcomes.setdefault(name, job.get("conclusion") or "unknown")

    for failure in failures:
        outcome = later_outcomes.get(failure.job_name)
        if outcome is None:
            continue
        failure.retried = True
        failure.retry_outcome = outcome


def ingest_run(
    client: GitHubClient,
    run: Dict[str, Any],
    attempt: Optional[int] = None,
    detect_flakes: bool = True,
) -> RunRecord:
    """Build a :class:`RunRecord` for one attempt of a workflow run."""
    run_id = run["id"]
    attempt = attempt or run.get("run_attempt", 1)

    jobs = client.list_run_jobs(run_id, attempt=attempt if attempt > 1 else None)
    counts, collected, failed_prepare = summarise_jobs(jobs)
    minutes = _accumulate_minutes(jobs)

    failures: List[JobFailure] = []
    for job in jobs:
        if job.get("conclusion") != "failure":
            continue
        parsed = parse_job_name(job.get("name", ""))
        failures.append(build_job_failure(job, parsed))

    if detect_flakes and run.get("run_attempt", 1) > attempt:
        annotate_retries(client, run_id, attempt, failures)

    pull_requests = run.get("pull_requests") or []
    record = RunRecord(
        run_id=run_id,
        attempt=attempt,
        workflow_name=run.get("name", ""),
        workflow_file=(run.get("path") or "").split("@")[0] or None,
        event=run.get("event"),
        head_branch=run.get("head_branch"),
        head_sha=run.get("head_sha"),
        pr_number=pull_requests[0]["number"] if pull_requests else None,
        created_at=run.get("created_at"),
        started_at=run.get("run_started_at") or run.get("created_at"),
        completed_at=run.get("updated_at"),
        conclusion=run.get("conclusion"),
        jobs_total=len(jobs),
        jobs_success=counts["success"],
        jobs_failure=counts["failure"],
        jobs_skipped=counts["skipped"],
        jobs_cancelled=counts["cancelled"],
        jobs_not_run_due_to_upstream=_estimate_lost_jobs(failed_prepare, collected),
        failed_prepare_jobs=[p.raw for p in failed_prepare],
        tests_collected_by_combo=collected,
        runner_minutes_total=round(minutes["runner_minutes_total"], 2),
        runner_minutes_failed=round(minutes["runner_minutes_failed"], 2),
        self_hosted_minutes=round(minutes["self_hosted_minutes"], 2),
        self_hosted_minutes_failed=round(minutes["self_hosted_minutes_failed"], 2),
        failures=failures,
        ingest_version=INGEST_VERSION,
        taxonomy_version=TAXONOMY_VERSION,
    )
    return record


def workflow_slug(record: RunRecord) -> str:
    """Filesystem-safe slug identifying the workflow a record belongs to."""
    source = record.workflow_file or record.workflow_name or "unknown"
    source = source.rsplit("/", 1)[-1]
    source = re.sub(r"\.ya?ml$", "", source)
    return re.sub(r"[^a-zA-Z0-9_.-]+", "-", source).strip("-").lower() or "unknown"


def record_month(record: RunRecord) -> str:
    timestamp = _parse_ts(record.created_at)
    if not timestamp:
        return "unknown"
    return timestamp.strftime("%Y-%m")
