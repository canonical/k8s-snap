#
# Copyright 2026 Canonical, Ltd.
#
"""Typed record shapes and SQLite schema for CI metrics."""

from __future__ import annotations

import dataclasses
import enum
from typing import Optional

SCHEMA_SQL = """
PRAGMA foreign_keys = ON;

CREATE TABLE schema_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);

CREATE TABLE workflows (
    workflow_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    file TEXT
);

CREATE TABLE configs (
    config_id INTEGER PRIMARY KEY,
    os        TEXT NOT NULL DEFAULT '',
    arch      TEXT NOT NULL DEFAULT '',
    channel   TEXT NOT NULL DEFAULT '',
    flavor    TEXT NOT NULL DEFAULT '',
    substrate TEXT NOT NULL DEFAULT '',
    UNIQUE (os, arch, channel, flavor, substrate)
);

CREATE TABLE tests (
    test_id INTEGER PRIMARY KEY,
    nodeid  TEXT NOT NULL UNIQUE,
    file    TEXT,
    owner   TEXT
);

CREATE TABLE runners (
    runner_id   INTEGER PRIMARY KEY,
    label_key   TEXT NOT NULL UNIQUE,
    self_hosted INTEGER NOT NULL
);
CREATE TABLE runner_labels (
    runner_id INTEGER NOT NULL REFERENCES runners ON DELETE CASCADE,
    label     TEXT NOT NULL,
    PRIMARY KEY (runner_id, label)
);

CREATE TABLE router_rules (
    router_rule_id TEXT PRIMARY KEY,
    failure_class  TEXT NOT NULL,
    subclass       TEXT,
    router_version INTEGER NOT NULL,
    CHECK (failure_class IN ('infra.runner','infra.provisioning',
        'external.dependency','ci.config','product.bug','test.bug','unknown'))
);

CREATE TABLE runs (
    run_id  INTEGER NOT NULL,
    attempt INTEGER NOT NULL,
    workflow_id INTEGER NOT NULL REFERENCES workflows,
    event TEXT, head_branch TEXT, head_sha TEXT, pr_number INTEGER,
    created_at TEXT, started_at TEXT, completed_at TEXT NOT NULL,
    conclusion TEXT NOT NULL,

    jobs_total     INTEGER NOT NULL,
    jobs_success   INTEGER NOT NULL,
    jobs_failure   INTEGER NOT NULL,
    jobs_skipped   INTEGER NOT NULL,
    jobs_cancelled INTEGER NOT NULL,
    runner_minutes_total  REAL NOT NULL DEFAULT 0,
    runner_minutes_failed REAL NOT NULL DEFAULT 0,

    ingest_version INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt)
);

CREATE TABLE run_tests_collected (
    run_id INTEGER NOT NULL, attempt INTEGER NOT NULL,
    config_id INTEGER NOT NULL REFERENCES configs,
    count INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt, config_id),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE
);

CREATE TABLE run_blocked_jobs (
    run_id INTEGER NOT NULL, attempt INTEGER NOT NULL,
    prepare_job_name TEXT NOT NULL,
    jobs_not_run INTEGER NOT NULL,
    PRIMARY KEY (run_id, attempt, prepare_job_name),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE
);

CREATE TABLE jobs (
    job_id  INTEGER PRIMARY KEY,
    run_id  INTEGER NOT NULL, attempt INTEGER NOT NULL,
    config_id INTEGER NOT NULL REFERENCES configs,
    test_id   INTEGER NOT NULL DEFAULT 0 REFERENCES tests,
    runner_id INTEGER REFERENCES runners,
    job_name TEXT NOT NULL, caller_job TEXT, html_url TEXT NOT NULL,
    started_at TEXT, completed_at TEXT, duration_s REAL,

    failed_step_name   TEXT,
    failed_step_number INTEGER,
    log_available      INTEGER,

    tested_revision TEXT NOT NULL DEFAULT '',
    tested_artifact TEXT NOT NULL DEFAULT '',
    collection_ref  TEXT NOT NULL DEFAULT '',
    collection_sha  TEXT NOT NULL DEFAULT '',

    signature_id   TEXT REFERENCES signatures,
    router_rule_id TEXT REFERENCES router_rules,
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE,
    CHECK (signature_id IS NULL OR router_rule_id IS NULL),
    CHECK (tested_revision = '' OR tested_artifact = '')
);

CREATE TABLE snap_provenance (
    channel  TEXT NOT NULL,
    arch     TEXT NOT NULL,
    revision INTEGER NOT NULL,
    built_sha TEXT NOT NULL,
    observed_at TEXT NOT NULL,
    PRIMARY KEY (channel, arch, revision)
);
CREATE INDEX jobs_by_signature ON jobs(signature_id);
CREATE INDEX jobs_by_run       ON jobs(run_id, attempt);
CREATE INDEX jobs_by_test      ON jobs(test_id);

CREATE TABLE signatures (
    signature_id TEXT PRIMARY KEY,
    normalised_pattern TEXT NOT NULL,
    exemplar_excerpt   TEXT NOT NULL,
    exemplar_job_id    INTEGER,
    normaliser_version INTEGER NOT NULL,
    extractor_version  INTEGER NOT NULL,
    state TEXT NOT NULL DEFAULT 'active',
    quarantined_at TEXT, quarantine_expires_at TEXT,
    CHECK (state IN ('active','quarantined','fixed','wontfix'))
);

CREATE TABLE signature_classifications (
    signature_id  TEXT NOT NULL REFERENCES signatures ON DELETE CASCADE,
    substrate     TEXT NOT NULL DEFAULT '*',
    ruleset_version INTEGER NOT NULL,
    failure_class TEXT NOT NULL,
    subclass      TEXT,
    rule_id       TEXT,
    confidence    REAL,
    classified_by TEXT NOT NULL,
    classified_at TEXT NOT NULL,
    PRIMARY KEY (signature_id, substrate, ruleset_version, classified_by),
    CHECK (failure_class IN ('infra.runner','infra.provisioning',
        'external.dependency','ci.config','product.bug','test.bug','unknown')),
    CHECK (classified_by IN ('rules','human'))
);

CREATE TABLE job_excerpts (
    job_id INTEGER PRIMARY KEY REFERENCES jobs ON DELETE CASCADE,
    excerpt TEXT NOT NULL,
    source  TEXT NOT NULL,
    extractor_version INTEGER NOT NULL
);
CREATE TABLE job_steps (
    job_id INTEGER NOT NULL REFERENCES jobs ON DELETE CASCADE,
    number INTEGER NOT NULL,
    name TEXT NOT NULL,
    conclusion TEXT,
    duration_s REAL,
    PRIMARY KEY (job_id, number)
);

CREATE TABLE retry_outcomes (
    run_id    INTEGER NOT NULL,
    attempt   INTEGER NOT NULL,
    config_id INTEGER NOT NULL REFERENCES configs,
    test_id   INTEGER NOT NULL DEFAULT 0 REFERENCES tests,
    conclusion TEXT NOT NULL,
    PRIMARY KEY (run_id, attempt, config_id, test_id),
    FOREIGN KEY (run_id, attempt) REFERENCES runs ON DELETE CASCADE,
    CHECK (conclusion IN ('success','failure','skipped','cancelled','not_run'))
);

CREATE TABLE issues (
    claim_id     INTEGER PRIMARY KEY,
    signature_id TEXT NOT NULL REFERENCES signatures ON DELETE CASCADE,
    substrate    TEXT NOT NULL DEFAULT '*',
    idempotency_key TEXT NOT NULL UNIQUE,
    issue_number INTEGER,
    issue_state  TEXT NOT NULL,
    claimed_at TEXT NOT NULL,
    opened_at TEXT, closed_at TEXT, closed_reason TEXT,
    fixed_at_sha TEXT,
    reopened_count INTEGER NOT NULL DEFAULT 0,
    post_attempts  INTEGER NOT NULL DEFAULT 0,
    last_comment_at TEXT,
    last_reported_occurrences INTEGER NOT NULL DEFAULT 0,
    occurrences_at_open INTEGER,
    opened_by_ruleset_version INTEGER,
    CHECK (issue_state IN ('claiming','open','closed','failed')),
    CHECK (issue_state IN ('claiming','failed') OR issue_number IS NOT NULL)
);
CREATE UNIQUE INDEX one_live_issue_per_signature
    ON issues(signature_id, substrate)
    WHERE issue_state IN ('claiming','open');
CREATE UNIQUE INDEX one_issue_number
    ON issues(issue_number) WHERE issue_number IS NOT NULL;
"""


class FailureClass(str, enum.Enum):
    """Axis A of the taxonomy: fault domain."""

    INFRA_RUNNER = "infra.runner"
    INFRA_PROVISIONING = "infra.provisioning"
    EXTERNAL_DEPENDENCY = "external.dependency"
    CI_CONFIG = "ci.config"
    PRODUCT_BUG = "product.bug"
    TEST_BUG = "test.bug"
    UNKNOWN = "unknown"


class ClassifiedBy(str, enum.Enum):
    """Stored classification provenance."""

    RULES = "rules"
    HUMAN = "human"


class SignatureState(str, enum.Enum):
    """Lifecycle of a durable signature."""

    ACTIVE = "active"
    QUARANTINED = "quarantined"
    FIXED = "fixed"
    WONTFIX = "wontfix"


@dataclasses.dataclass(frozen=True)
class Workflow:
    workflow_id: int
    name: str
    file: Optional[str] = None


@dataclasses.dataclass(frozen=True)
class Config:
    config_id: int
    os: str = ""
    arch: str = ""
    channel: str = ""
    flavor: str = ""
    substrate: str = ""


@dataclasses.dataclass(frozen=True)
class Test:
    test_id: int
    nodeid: str
    file: Optional[str] = None
    owner: Optional[str] = None


@dataclasses.dataclass(frozen=True)
class Runner:
    runner_id: int
    label_key: str
    self_hosted: int


@dataclasses.dataclass(frozen=True)
class RunnerLabel:
    runner_id: int
    label: str


@dataclasses.dataclass(frozen=True)
class RouterRule:
    router_rule_id: str
    failure_class: str
    subclass: Optional[str]
    router_version: int


@dataclasses.dataclass(frozen=True)
class Run:
    run_id: int
    attempt: int
    workflow_id: int
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
    ingest_version: int


@dataclasses.dataclass(frozen=True)
class RunTestsCollected:
    run_id: int
    attempt: int
    config_id: int
    count: int


@dataclasses.dataclass(frozen=True)
class RunBlockedJob:
    run_id: int
    attempt: int
    prepare_job_name: str
    jobs_not_run: int


@dataclasses.dataclass(frozen=True)
class Job:
    job_id: int
    run_id: int
    attempt: int
    config_id: int
    test_id: int
    runner_id: Optional[int]
    job_name: str
    caller_job: Optional[str]
    html_url: str
    started_at: Optional[str]
    completed_at: Optional[str]
    duration_s: Optional[float]
    failed_step_name: Optional[str]
    failed_step_number: Optional[int]
    log_available: Optional[int]
    tested_revision: str = ""
    tested_artifact: str = ""
    collection_ref: str = ""
    collection_sha: str = ""
    signature_id: Optional[str] = None
    router_rule_id: Optional[str] = None


@dataclasses.dataclass(frozen=True)
class SnapProvenance:
    channel: str
    arch: str
    revision: int
    built_sha: str
    observed_at: str


@dataclasses.dataclass(frozen=True)
class Signature:
    signature_id: str
    normalised_pattern: str
    exemplar_excerpt: str
    exemplar_job_id: Optional[int]
    normaliser_version: int
    extractor_version: int
    state: str = SignatureState.ACTIVE.value
    quarantined_at: Optional[str] = None
    quarantine_expires_at: Optional[str] = None


@dataclasses.dataclass(frozen=True)
class SignatureClassification:
    signature_id: str
    substrate: str
    ruleset_version: int
    failure_class: str
    subclass: Optional[str]
    rule_id: Optional[str]
    confidence: Optional[float]
    classified_by: str
    classified_at: str


@dataclasses.dataclass(frozen=True)
class JobExcerpt:
    job_id: int
    excerpt: str
    source: str
    extractor_version: int


@dataclasses.dataclass(frozen=True)
class JobStep:
    job_id: int
    number: int
    name: str
    conclusion: Optional[str]
    duration_s: Optional[float]


@dataclasses.dataclass(frozen=True)
class RetryOutcome:
    run_id: int
    attempt: int
    config_id: int
    test_id: int
    conclusion: str


@dataclasses.dataclass(frozen=True)
class Issue:
    claim_id: int
    signature_id: str
    substrate: str
    idempotency_key: str
    issue_number: Optional[int]
    issue_state: str
    claimed_at: str
    opened_at: Optional[str]
    closed_at: Optional[str]
    closed_reason: Optional[str]
    fixed_at_sha: Optional[str]
    reopened_count: int
    post_attempts: int
    last_comment_at: Optional[str]
    last_reported_occurrences: int
    occurrences_at_open: Optional[int]
    opened_by_ruleset_version: Optional[int]
