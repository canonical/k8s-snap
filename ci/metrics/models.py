#
# Copyright 2026 Canonical, Ltd.
#
"""Record shapes for the CI metrics store.

Three record families are defined here:

* :class:`RunRecord` -- one per workflow run attempt, always written.
* :class:`JobFailure` -- one per failed job. Passing jobs are *not* stored
  individually; they are reduced to counts on the run record. This is what
  keeps the committed data volume small.
* :class:`Signature` -- the durable unit of work. A signature groups every
  job failure sharing the same normalised error fingerprint, so that the
  36 nightly jobs failing on one root cause become one thing to own and fix.

Every record carries the versions of the code that produced it
(``ingest_version``, ``taxonomy_version``, ``ruleset_version``,
``normaliser_version``) so that historical comparisons remain honest when
the pipeline changes underneath them.
"""

import dataclasses
import enum
from typing import Any, Dict, List, Optional


class FailureClass(str, enum.Enum):
    """Axis A of the taxonomy: fault domain, i.e. who owns the fix.

    Ordering matters. The classifier walks classes in declaration order and
    takes the first match, so an infrastructure failure that happens to also
    print a product error is not blamed on the product.
    """

    INFRA_RUNNER = "infra.runner"
    INFRA_PROVISIONING = "infra.provisioning"
    EXTERNAL_DEPENDENCY = "external.dependency"
    CI_CONFIG = "ci.config"
    PRODUCT_BUG = "product.bug"
    TEST_BUG = "test.bug"
    UNKNOWN = "unknown"


# Permitted subclasses per class. The classifier validates against this map so
# a typo in the rule pack fails loudly instead of silently inventing a category.
SUBCLASSES: Dict[FailureClass, List[str]] = {
    FailureClass.INFRA_RUNNER: [
        "runner_lost",
        "job_timeout",
        "disk_space",
        "oom",
        "self_hosted_unavailable",
        "node_cleanup",
        # A tool the workflow assumes exists is absent from the image.
        "missing_tool",
    ],
    FailureClass.INFRA_PROVISIONING: [
        "lxd_setup",
        "multipass_setup",
        "image_pull",
        "bridge_create",
        "instance_launch",
    ],
    FailureClass.EXTERNAL_DEPENDENCY: [
        "snap_store",
        "registry",
        "apt",
        "pypi",
        "launchpad",
        "github_api",
        "dns",
    ],
    FailureClass.CI_CONFIG: [
        "test_collection",
        "bad_input",
        "missing_artifact",
        "secret_missing",
        "workflow_logic",
        "tox_env",
    ],
    FailureClass.PRODUCT_BUG: [
        "service_not_active",
        "panic",
        "api_error",
        "upgrade_failure",
        "networking",
        "storage",
        # Added in taxonomy v2 after the 90-day backfill showed that the
        # integration harness funnels nearly every product failure through
        # one CalledProcessError frame, leaving the wrapped command as the
        # only discriminator. These name the four commands that account for
        # the bulk of it, so "product.bug" stops being a single opaque heap.
        "wait_for_timeout",
        "cluster_never_ready",
        "bootstrap_failure",
        "snap_install_failure",
        "join_failure",
    ],
    FailureClass.TEST_BUG: [
        "timeout_too_short",
        "bad_assumption",
        "resource_leak",
        "race_in_test",
        "unmaintained",
        # The harness itself raised, so no product code ran at all.
        "harness_bug",
    ],
    FailureClass.UNKNOWN: [],
}


class Reproducibility(str, enum.Enum):
    """Axis B: derived from history, never from a single job."""

    FLAKY = "flaky"
    SYSTEMIC = "systemic"
    INTERMITTENT = "intermittent"
    NEW = "new"


class ClassifiedBy(str, enum.Enum):
    """Provenance of a classification verdict.

    ``LLM`` never appears on a stored record: model output is emitted as a
    proposed rule-pack change for human review, and only becomes a recorded
    verdict once merged (at which point provenance is ``RULES``).
    """

    ROUTER = "router"
    RULES = "rules"
    HUMAN = "human"
    NONE = "none"


class SignatureState(str, enum.Enum):
    ACTIVE = "active"
    QUARANTINED = "quarantined"
    FIXED = "fixed"
    WONTFIX = "wontfix"


def _asdict_enum(obj: Any) -> Any:
    """Recursively convert dataclasses to plain JSON-serialisable dicts."""
    if isinstance(obj, enum.Enum):
        return obj.value
    if dataclasses.is_dataclass(obj) and not isinstance(obj, type):
        return {k: _asdict_enum(v) for k, v in dataclasses.asdict(obj).items()}
    if isinstance(obj, dict):
        return {k: _asdict_enum(v) for k, v in obj.items()}
    if isinstance(obj, (list, tuple)):
        return [_asdict_enum(v) for v in obj]
    return obj


@dataclasses.dataclass
class StepRecord:
    """A single step of a job, reduced to what classification needs."""

    name: str
    number: int
    conclusion: Optional[str]
    duration_s: Optional[float] = None

    def to_dict(self) -> Dict[str, Any]:
        return _asdict_enum(self)


STEP_FIELDS = {f.name for f in dataclasses.fields(StepRecord)}


@dataclasses.dataclass
class JobFailure:
    """One failed job, with enough evidence to reclassify it later offline.

    ``evidence_excerpt`` holds the scrubbed, normalised error text rather than
    the full log. That is the key enabler for ``metrics reclassify``: when the
    rule pack or the normaliser changes, history can be recomputed without
    refetching a single log.
    """

    job_id: int
    run_id: int
    attempt: int
    job_name: str
    html_url: str

    # Dimensions parsed out of the job name (see ingest.parse_job_name).
    caller_job: Optional[str] = None
    os: Optional[str] = None
    arch: Optional[str] = None
    channel: Optional[str] = None
    flavor: Optional[str] = None
    substrate: Optional[str] = None
    test_nodeid: Optional[str] = None
    test_file: Optional[str] = None

    runner_labels: List[str] = dataclasses.field(default_factory=list)
    self_hosted: bool = False

    started_at: Optional[str] = None
    completed_at: Optional[str] = None
    duration_s: Optional[float] = None

    failed_step_name: Optional[str] = None
    failed_step_number: Optional[int] = None
    steps: List[StepRecord] = dataclasses.field(default_factory=list)
    log_available: Optional[bool] = None

    # Populated in P1.4 (signatures) and P1.3/P1.5 (classification).
    signature_id: Optional[str] = None
    evidence_excerpt: Optional[str] = None
    failure_class: str = FailureClass.UNKNOWN.value
    subclass: Optional[str] = None
    confidence: Optional[float] = None
    classified_by: str = ClassifiedBy.NONE.value
    classifier_version: Optional[str] = None
    normaliser_version: Optional[int] = None
    ruleset_version: Optional[int] = None
    rule_id: Optional[str] = None

    reproducibility: Optional[str] = None
    retried: bool = False
    retry_outcome: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        return _asdict_enum(self)

    @classmethod
    def from_dict(cls, payload: Dict[str, Any]) -> "JobFailure":
        """Rebuild a failure from a stored record.

        Unknown keys are dropped rather than raising, so a rule pack can be
        replayed over records written by an older schema version. That
        tolerance is the whole point of retaining excerpts: reclassification
        must keep working across schema churn, otherwise history becomes
        read-only the first time a field is added.
        """
        fields = {f.name for f in dataclasses.fields(cls)}
        kwargs = {k: v for k, v in payload.items() if k in fields}
        kwargs["steps"] = [
            StepRecord(**{k: v for k, v in step.items() if k in STEP_FIELDS})
            for step in payload.get("steps") or []
        ]
        return cls(**kwargs)


@dataclasses.dataclass
class RunRecord:
    """One workflow run attempt, plus aggregate counts for its jobs."""

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
    completed_at: Optional[str]
    conclusion: Optional[str]

    jobs_total: int = 0
    jobs_success: int = 0
    jobs_failure: int = 0
    jobs_skipped: int = 0
    jobs_cancelled: int = 0

    # Blind-spot counter. When a `Prepare Environment` job fails, its whole
    # downstream test matrix never runs and therefore never appears as failed.
    # Without this, a prepare failure makes the metrics look *better*.
    jobs_not_run_due_to_upstream: int = 0
    failed_prepare_jobs: List[str] = dataclasses.field(default_factory=list)

    # Denominator integrity: test counts vary per (os, arch, channel) because
    # each channel collects tests from its matching release branch.
    tests_collected_by_combo: Dict[str, int] = dataclasses.field(default_factory=dict)

    runner_minutes_total: float = 0.0
    runner_minutes_failed: float = 0.0
    self_hosted_minutes: float = 0.0
    self_hosted_minutes_failed: float = 0.0

    failures: List[JobFailure] = dataclasses.field(default_factory=list)

    ingest_version: int = 0
    taxonomy_version: int = 0
    ruleset_version: Optional[int] = None

    def to_dict(self) -> Dict[str, Any]:
        return _asdict_enum(self)


@dataclasses.dataclass
class Signature:
    """A durable failure fingerprint: the unit we count, own, and fix."""

    signature_id: str
    normalised_pattern: str
    exemplar_excerpt: str
    exemplar_job_url: str

    failure_class: str = FailureClass.UNKNOWN.value
    subclass: Optional[str] = None
    owner: Optional[str] = None
    linked_issue: Optional[int] = None

    first_seen: Optional[str] = None
    last_seen: Optional[str] = None
    runs_seen: int = 0
    total_occurrences: int = 0

    state: str = SignatureState.ACTIVE.value
    quarantined_at: Optional[str] = None
    quarantine_expires_at: Optional[str] = None

    reproducibility: Optional[str] = None
    configs_affected: List[str] = dataclasses.field(default_factory=list)

    def to_dict(self) -> Dict[str, Any]:
        return _asdict_enum(self)
