#
# Copyright 2026 Canonical, Ltd.
#
"""Stage A: deterministic, log-free failure routing.

The router answers "who owns this failure?" using **only** job and step
metadata that the jobs endpoint already returned. It fetches nothing. That
matters for two reasons:

1. **Cost.** A nightly run has ~2,500 jobs. Routing them costs zero extra API
   calls, so the infra/external/product split for a whole run is available
   immediately after ingest.
2. **Coverage of the log-less cases.** A meaningful share of failures have no
   log to read at all -- see :func:`_route_no_failed_step`. Any classifier
   built purely on log text scores these ``unknown`` forever.

The router is intentionally *incomplete*. It resolves the failures whose cause
is unambiguous from the step that failed -- infrastructure, provisioning and
external dependencies -- and explicitly defers anything that failed inside a
test step to Stage B, where the log excerpt and rule pack decide. Deferring is
a first-class outcome, not a fallback: a confident wrong answer here would
mis-attribute product bugs to infrastructure and quietly corrupt the baseline.

Ordering is significant. Rules are evaluated top to bottom and the first match
wins, so the more specific patterns must precede the general ones.
"""

import re
from typing import Any, Callable, Dict, List, NamedTuple, Optional, Tuple

from metrics.models import ClassifiedBy, FailureClass, JobFailure

ROUTER_VERSION = 1

# Marker meaning "metadata is not enough, Stage B must look at the log".
DEFER = "defer"


class Verdict(NamedTuple):
    """Outcome of routing one job."""

    failure_class: Optional[str]
    subclass: Optional[str]
    confidence: float
    rule_id: str
    defer: bool = False


class StepRule(NamedTuple):
    """Maps a failed step name to a fault domain.

    ``guard`` narrows a rule to jobs satisfying an extra metadata predicate.
    Step names are not always unambiguous: the same composite action can do
    two unrelated things depending on its inputs, and the step name is
    identical either way. A guard lets the fault domain follow the input
    rather than the label.
    """

    rule_id: str
    pattern: str
    failure_class: Optional[str]
    subclass: Optional[str]
    confidence: float
    defer: bool = False
    guard: Optional[Callable[[JobFailure], bool]] = None


# Ordered. First match wins.
STEP_RULES: Tuple[StepRule, ...] = (
    # -- deferrals first -------------------------------------------------
    # A test step that failed says nothing about *why*: the same
    # `Run test_version_upgrades` step covers a genuine upgrade regression, a
    # flaky wait, and a snap-store 503. Only the log can separate them.
    StepRule("defer.test-step", r"^Run\s+test_", None, None, 0.0, defer=True),
    StepRule("defer.pytest", r"^Run\s+(pytest|tox)\b", None, None, 0.0, defer=True),
    # -- external dependencies -------------------------------------------
    # `Download k8s-snap` is a composite action with two mutually exclusive
    # modes, and the step name is the same in both:
    #
    #   channel mode  -> `snap download k8s --channel=...` -- the snap store
    #   artifact mode -> actions/download-artifact -- *our own* build output
    #
    # Only the first is external. Conflating them would file every missing
    # build artifact under "external -- usually wait/retry", telling the team
    # to ignore a problem they own. The job name carries the channel for
    # matrix jobs, so the mode is recoverable from metadata alone.
    StepRule(
        "external.snap-download",
        r"^(Download|Fetch|Install)\s+k8s-snap\b",
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "snap_store",
        0.9,
        guard=lambda job: bool(job.channel),
    ),
    StepRule(
        "ci.snap-artifact-missing",
        r"^(Download|Fetch|Install)\s+k8s-snap\b",
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
        0.8,
    ),
    StepRule(
        "external.snapcraft-login",
        r"snapcraft\s+(login|whoami|upload|release)",
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "snap_store",
        0.85,
    ),
    StepRule(
        "external.checkout",
        r"^(Checkout|actions/checkout)\b",
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "github_api",
        0.8,
    ),
    StepRule(
        "external.download-artifact",
        r"^Download\b.*\bartifact\b",
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
        0.75,
    ),
    StepRule(
        "external.registry",
        r"\b(docker|ghcr|registry|image pull|skopeo)\b",
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "registry",
        0.7,
    ),
    StepRule(
        "external.apt",
        r"\bapt(-get)?\b|Install\s+dependencies",
        FailureClass.EXTERNAL_DEPENDENCY.value,
        "apt",
        0.65,
    ),
    # -- provisioning ----------------------------------------------------
    StepRule(
        "provisioning.lxd",
        r"\b(Setup|Install|Initialise|Initialize|Configure)\s+LXD\b",
        FailureClass.INFRA_PROVISIONING.value,
        "lxd_setup",
        0.9,
    ),
    StepRule(
        "provisioning.multipass",
        r"\b(Setup|Install)\s+Multipass\b",
        FailureClass.INFRA_PROVISIONING.value,
        "multipass_setup",
        0.9,
    ),
    StepRule(
        "provisioning.prepare-env",
        r"^Prepare\s+Environment\b",
        FailureClass.INFRA_PROVISIONING.value,
        "instance_launch",
        0.8,
    ),
    StepRule(
        "provisioning.image",
        r"\b(Build|Import|Pull)\s+(image|rootfs)\b",
        FailureClass.INFRA_PROVISIONING.value,
        "image_pull",
        0.7,
    ),
    # -- runner ----------------------------------------------------------
    StepRule(
        "runner.setup-job",
        r"^Set up job$",
        FailureClass.INFRA_RUNNER.value,
        "self_hosted_unavailable",
        0.9,
    ),
    StepRule(
        "runner.cleanup",
        r"^(Node Cleanup|Cleanup|Complete job)$",
        FailureClass.INFRA_RUNNER.value,
        "node_cleanup",
        0.85,
    ),
    StepRule(
        "runner.post-step",
        r"^Post\s+",
        FailureClass.INFRA_RUNNER.value,
        "node_cleanup",
        0.7,
    ),
    StepRule(
        "runner.disk",
        r"\b(Free disk space|Maximize build space|Prune)\b",
        FailureClass.INFRA_RUNNER.value,
        "disk_space",
        0.75,
    ),
    # -- CI configuration -------------------------------------------------
    StepRule(
        "ci.collect-tests",
        r"^(Collect|Get)\s+.*tests?\b|^Get e2e test tags$",
        FailureClass.CI_CONFIG.value,
        "test_collection",
        0.85,
    ),
    StepRule(
        "ci.upload-artifact",
        r"^Upload\b",
        FailureClass.CI_CONFIG.value,
        "missing_artifact",
        0.7,
    ),
    StepRule(
        "ci.tox",
        r"\btox\b|\bvirtualenv\b|\bpip install\b",
        FailureClass.CI_CONFIG.value,
        "tox_env",
        0.6,
    ),
)

_COMPILED: List[Tuple[StepRule, "re.Pattern[str]"]] = [
    (rule, re.compile(rule.pattern, re.IGNORECASE)) for rule in STEP_RULES
]


def _route_no_failed_step(job: JobFailure) -> Optional[Verdict]:
    """Classify a job that failed without any step reporting failure.

    This is the case a log-based classifier cannot see. The job concludes
    ``failure``, every step that ran concluded ``success``, the remaining steps
    are left ``null`` (they never got to run), and GitHub serves no log blob at
    all -- the API returns 404 BlobNotFound. That combination means the runner
    disappeared underneath the job; nothing about the product or the test was
    ever demonstrated.

    Verified against real job 101894558721 (nightly run 34172139128): all setup
    steps green, ``Run test_...`` and everything after it ``null``, 52 minutes
    elapsed, no log.

    A cancelled or timed-out job reaches the same shape, so the two are
    separated by duration against the GitHub 6-hour job ceiling.
    """
    if job.failed_step_name is not None:
        return None

    trailing_null = any(step.conclusion is None for step in job.steps)
    log_missing = job.log_available is False

    if job.duration_s is not None and job.duration_s >= 6 * 3600 - 300:
        return Verdict(
            FailureClass.INFRA_RUNNER.value, "job_timeout", 0.9, "runner.job-timeout"
        )

    if trailing_null or log_missing or not job.steps:
        # Confidence is deliberately high: no other mechanism produces this
        # exact shape, and mislabelling it as `unknown` is the failure mode we
        # are trying to eliminate.
        confidence = 0.9 if (trailing_null and log_missing) else 0.75
        return Verdict(
            FailureClass.INFRA_RUNNER.value, "runner_lost", confidence, "runner.lost"
        )

    return None


def route(job: JobFailure) -> Verdict:
    """Classify one failed job from metadata alone.

    Returns a deferring verdict when the failed step is a test step, meaning
    Stage B must decide. Never returns ``unknown`` for a resolvable case.
    """
    no_step = _route_no_failed_step(job)
    if no_step is not None:
        return no_step

    step_name = job.failed_step_name or ""
    for rule, compiled in _COMPILED:
        if not compiled.search(step_name):
            continue
        if rule.guard is not None and not rule.guard(job):
            continue
        if rule.defer:
            return Verdict(None, None, 0.0, rule.rule_id, defer=True)
        return Verdict(rule.failure_class, rule.subclass, rule.confidence, rule.rule_id)

    # An unrecognised non-test step is still more likely CI plumbing than a
    # product bug, but not confidently enough to record it. Defer.
    return Verdict(None, None, 0.0, "defer.unmatched-step", defer=True)


def apply_router(job: JobFailure) -> Verdict:
    """Route a job and write the verdict onto it in place."""
    verdict = route(job)
    if not verdict.defer:
        job.failure_class = verdict.failure_class or FailureClass.UNKNOWN.value
        job.subclass = verdict.subclass
        job.confidence = verdict.confidence
        job.classified_by = ClassifiedBy.ROUTER.value
        job.classifier_version = f"router-v{ROUTER_VERSION}"
        job.rule_id = verdict.rule_id
    return verdict


def route_all(jobs: List[JobFailure]) -> Dict[str, Any]:
    """Route a whole run's failures and summarise the outcome."""
    by_class: Dict[str, int] = {}
    deferred = 0
    for job in jobs:
        verdict = apply_router(job)
        if verdict.defer:
            deferred += 1
            continue
        key = f"{verdict.failure_class}/{verdict.subclass}"
        by_class[key] = by_class.get(key, 0) + 1
    return {
        "router_version": ROUTER_VERSION,
        "total": len(jobs),
        "routed": len(jobs) - deferred,
        "deferred_to_stage_b": deferred,
        "by_class": dict(sorted(by_class.items(), key=lambda kv: -kv[1])),
    }
