#
# Copyright 2026 Canonical, Ltd.
#
"""Stage A metadata-only failure router."""

from __future__ import annotations

import dataclasses
import re
import sqlite3
from typing import Callable, Dict, Iterable, Optional, Pattern, Tuple

from metrics.models import FailureClass

ROUTER_VERSION = 1


@dataclasses.dataclass(frozen=True)
class RouterJob:
    """The durable job metadata Stage A is allowed to inspect."""

    job_id: int
    failed_step_name: Optional[str]
    log_available: Optional[int]
    duration_s: Optional[float] = None
    channel: str = ""


@dataclasses.dataclass(frozen=True)
class RouterVerdict:
    """A metadata-router verdict."""

    router_rule_id: str
    failure_class: Optional[str]
    subclass: Optional[str]
    defer: bool


@dataclasses.dataclass(frozen=True)
class StepRule:
    """Maps a failed-step name to a router verdict."""

    router_rule_id: str
    pattern: Pattern[str]
    failure_class: Optional[str]
    subclass: Optional[str]
    defer: bool = False
    guard: Optional[Callable[[RouterJob], bool]] = None


DEFER_TEST_STEP = RouterVerdict("defer.test-step", None, None, True)
DEFER_UNMATCHED_STEP = RouterVerdict("defer.unmatched-step", None, None, True)

STEP_RULES: Tuple[StepRule, ...] = (
    StepRule(
        "defer.test-step",
        re.compile(r"^Run\s+test_", re.IGNORECASE),
        None,
        None,
        defer=True,
    ),
    StepRule(
        "defer.pytest",
        re.compile(r"^Run\s+(pytest|tox)\b", re.IGNORECASE),
        None,
        None,
        defer=True,
    ),
    StepRule(
        "external.snap-download",
        re.compile(r"^(Download|Fetch|Install)\s+k8s-snap\b", re.IGNORECASE),
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "snap_store",
        guard=lambda job: bool(job.channel),
    ),
    StepRule(
        "ci.snap-artifact-missing",
        re.compile(r"^(Download|Fetch|Install)\s+k8s-snap\b", re.IGNORECASE),
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
    ),
    StepRule(
        "external.checkout",
        re.compile(r"^(Checkout|actions/checkout)\b", re.IGNORECASE),
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "github_api",
    ),
    StepRule(
        "ci.download-artifact",
        re.compile(r"^Download\b.*\bartifact\b", re.IGNORECASE),
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
    ),
    StepRule(
        "external.registry",
        re.compile(r"\b(docker|ghcr|registry|image pull|skopeo)\b", re.IGNORECASE),
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "registry",
    ),
    StepRule(
        "external.apt",
        re.compile(r"\bapt(-get)?\b|Install\s+dependencies", re.IGNORECASE),
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "apt",
    ),
    StepRule(
        "infra.provisioning.lxd",
        re.compile(
            r"\b(Setup|Install|Initialise|Initialize|Configure)\s+LXD\b",
            re.IGNORECASE,
        ),
        FailureClass.INFRA_PROVISIONING.value,
        "lxd_setup",
    ),
    StepRule(
        "infra.provisioning.multipass",
        re.compile(r"\b(Setup|Install)\s+Multipass\b", re.IGNORECASE),
        FailureClass.INFRA_PROVISIONING.value,
        "multipass_setup",
    ),
    StepRule(
        "infra.provisioning.prepare-env",
        re.compile(r"^Prepare\s+Environment\b", re.IGNORECASE),
        FailureClass.INFRA_PROVISIONING.value,
        "instance_launch",
    ),
    StepRule(
        "infra.runner.setup-job",
        re.compile(r"^Set up job$", re.IGNORECASE),
        FailureClass.INFRA_RUNNER.value,
        "self_hosted_unavailable",
    ),
    StepRule(
        "infra.runner.cleanup",
        re.compile(r"^(Node Cleanup|Cleanup|Complete job)$", re.IGNORECASE),
        FailureClass.INFRA_RUNNER.value,
        "node_cleanup",
    ),
    StepRule(
        "infra.runner.post-step",
        re.compile(r"^Post\s+", re.IGNORECASE),
        FailureClass.INFRA_RUNNER.value,
        "node_cleanup",
    ),
    StepRule(
        "infra.runner.disk",
        re.compile(r"\b(Free disk space|Maximize build space|Prune)\b", re.IGNORECASE),
        FailureClass.INFRA_RUNNER.value,
        "disk_space",
    ),
    StepRule(
        "ci.collect-tests",
        re.compile(r"^(Collect|Get)\s+.*tests?\b|^Get e2e test tags$", re.IGNORECASE),
        FailureClass.CI_CONFIG.value,
        "test_collection",
    ),
    StepRule(
        "ci.upload-artifact",
        re.compile(r"^Upload\b", re.IGNORECASE),
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
    ),
)


def router_rules() -> Tuple[Tuple[str, str, Optional[str], int], ...]:
    """Return the router rule dimension rows."""

    rows = [
        (
            "infra.runner.lost",
            FailureClass.INFRA_RUNNER.value,
            "runner_lost",
            ROUTER_VERSION,
        ),
        (
            "infra.runner.job-timeout",
            FailureClass.INFRA_RUNNER.value,
            "job_timeout",
            ROUTER_VERSION,
        ),
    ]
    rows.extend(
        (
            rule.router_rule_id,
            rule.failure_class,
            rule.subclass,
            ROUTER_VERSION,
        )
        for rule in STEP_RULES
        if not rule.defer and rule.failure_class is not None
    )
    return tuple(rows)


def route(job: RouterJob) -> RouterVerdict:
    """Classify one failed job from metadata alone, or defer to Stage B."""

    if job.failed_step_name is None:
        if job.duration_s is not None and job.duration_s >= 6 * 3600 - 300:
            return RouterVerdict(
                "infra.runner.job-timeout",
                FailureClass.INFRA_RUNNER.value,
                "job_timeout",
                False,
            )
        if job.log_available == 0:
            return RouterVerdict(
                "infra.runner.lost",
                FailureClass.INFRA_RUNNER.value,
                "runner_lost",
                False,
            )
        return DEFER_UNMATCHED_STEP

    for rule in STEP_RULES:
        if not rule.pattern.search(job.failed_step_name):
            continue
        if rule.guard is not None and not rule.guard(job):
            continue
        return RouterVerdict(
            rule.router_rule_id,
            rule.failure_class,
            rule.subclass,
            rule.defer,
        )
    return DEFER_UNMATCHED_STEP


def ensure_router_rules(conn: sqlite3.Connection) -> None:
    """Insert the router rule lookup rows used by ``v_job_class``."""

    conn.executemany(
        """
        INSERT INTO router_rules (
            router_rule_id, failure_class, subclass, router_version
        ) VALUES (?, ?, ?, ?)
        ON CONFLICT(router_rule_id) DO UPDATE SET
            failure_class = excluded.failure_class,
            subclass = excluded.subclass,
            router_version = excluded.router_version
        """,
        router_rules(),
    )


def _jobs(conn: sqlite3.Connection) -> Iterable[RouterJob]:
    rows = conn.execute(
        """
        SELECT j.job_id, j.failed_step_name, j.log_available, j.duration_s, cfg.channel
        FROM jobs j
        JOIN configs cfg USING (config_id)
        ORDER BY j.job_id
        """
    )
    for row in rows:
        yield RouterJob(
            job_id=int(row[0]),
            failed_step_name=row[1],
            log_available=row[2],
            duration_s=row[3],
            channel=row[4] or "",
        )


def apply_router(conn: sqlite3.Connection) -> Dict[str, int]:
    """Apply Stage A to stored jobs and return a compact summary."""

    ensure_router_rules(conn)
    summary = {"routed": 0, "deferred": 0}
    by_rule: Dict[str, int] = {}

    for job in _jobs(conn):
        verdict = route(job)
        if verdict.defer:
            summary["deferred"] += 1
            continue
        conn.execute(
            """
            UPDATE jobs
            SET router_rule_id = ?, signature_id = NULL
            WHERE job_id = ?
            """,
            (verdict.router_rule_id, job.job_id),
        )
        summary["routed"] += 1
        by_rule[verdict.router_rule_id] = by_rule.get(verdict.router_rule_id, 0) + 1

    return {
        **summary,
        **{f"rule:{key}": value for key, value in sorted(by_rule.items())},
    }
