#
# Copyright 2026 Canonical, Ltd.
#
"""SQLite persistence and deterministic text dumps for CI metrics."""

from __future__ import annotations

import hashlib
import os
import sqlite3
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterable, Optional, Sequence, Tuple

from . import SCHEMA_VERSION
from .models import SCHEMA_SQL

EPHEMERAL_TABLES = ("job_excerpts", "job_steps")
DURABLE_TABLES = (
    "schema_meta",
    "workflows",
    "configs",
    "tests",
    "runners",
    "runner_labels",
    "router_rules",
    "signatures",
    "runs",
    "run_tests_collected",
    "run_blocked_jobs",
    "jobs",
    "snap_provenance",
    "signature_classifications",
    "retry_outcomes",
    "issues",
)


@dataclass(frozen=True)
class DumpTable:
    name: str
    columns: Tuple[str, ...]
    select_sql: str


class InFlightRunError(ValueError):
    """Raised when a run cannot be frozen because it has no conclusion."""


DUMP_TABLES = (
    DumpTable(
        "schema_meta",
        ("key", "value"),
        "SELECT key, value FROM schema_meta ORDER BY key",
    ),
    DumpTable(
        "workflows",
        ("workflow_id", "name", "file"),
        "SELECT workflow_id, name, file FROM workflows ORDER BY name, workflow_id",
    ),
    DumpTable(
        "configs",
        ("config_id", "os", "arch", "channel", "flavor", "substrate"),
        """
        SELECT config_id, os, arch, channel, flavor, substrate
        FROM configs ORDER BY os, arch, channel, flavor, substrate, config_id
        """,
    ),
    DumpTable(
        "tests",
        ("test_id", "nodeid", "file", "owner"),
        "SELECT test_id, nodeid, file, owner FROM tests ORDER BY nodeid, test_id",
    ),
    DumpTable(
        "runners",
        ("runner_id", "label_key", "self_hosted"),
        """
        SELECT runner_id, label_key, self_hosted
        FROM runners ORDER BY label_key, runner_id
        """,
    ),
    DumpTable(
        "runner_labels",
        ("runner_id", "label"),
        "SELECT runner_id, label FROM runner_labels ORDER BY runner_id, label",
    ),
    DumpTable(
        "router_rules",
        ("router_rule_id", "failure_class", "subclass", "router_version"),
        """
        SELECT router_rule_id, failure_class, subclass, router_version
        FROM router_rules ORDER BY router_rule_id
        """,
    ),
    DumpTable(
        "signatures",
        (
            "signature_id",
            "normalised_pattern",
            "exemplar_excerpt",
            "exemplar_job_id",
            "normaliser_version",
            "extractor_version",
            "state",
            "quarantined_at",
            "quarantine_expires_at",
        ),
        """
        SELECT signature_id, normalised_pattern, exemplar_excerpt, exemplar_job_id,
               normaliser_version, extractor_version, state, quarantined_at,
               quarantine_expires_at
        FROM signatures ORDER BY signature_id
        """,
    ),
    DumpTable(
        "runs",
        (
            "run_id",
            "attempt",
            "workflow_id",
            "event",
            "head_branch",
            "head_sha",
            "pr_number",
            "created_at",
            "started_at",
            "completed_at",
            "conclusion",
            "jobs_total",
            "jobs_success",
            "jobs_failure",
            "jobs_skipped",
            "jobs_cancelled",
            "runner_minutes_total",
            "runner_minutes_failed",
            "ingest_version",
        ),
        """
        SELECT run_id, attempt, workflow_id, event, head_branch, head_sha,
               pr_number, created_at, started_at, completed_at, conclusion,
               jobs_total, jobs_success, jobs_failure, jobs_skipped,
               jobs_cancelled, runner_minutes_total, runner_minutes_failed,
               ingest_version
        FROM runs ORDER BY run_id, attempt
        """,
    ),
    DumpTable(
        "run_tests_collected",
        ("run_id", "attempt", "config_id", "count"),
        """
        SELECT run_id, attempt, config_id, count
        FROM run_tests_collected ORDER BY run_id, attempt, config_id
        """,
    ),
    DumpTable(
        "run_blocked_jobs",
        ("run_id", "attempt", "prepare_job_name", "jobs_not_run"),
        """
        SELECT run_id, attempt, prepare_job_name, jobs_not_run
        FROM run_blocked_jobs ORDER BY run_id, attempt, prepare_job_name
        """,
    ),
    DumpTable(
        "jobs",
        (
            "job_id",
            "run_id",
            "attempt",
            "config_id",
            "test_id",
            "runner_id",
            "job_name",
            "caller_job",
            "html_url",
            "started_at",
            "completed_at",
            "duration_s",
            "failed_step_name",
            "failed_step_number",
            "log_available",
            "tested_revision",
            "tested_artifact",
            "collection_ref",
            "collection_sha",
            "signature_id",
            "router_rule_id",
        ),
        """
        SELECT job_id, run_id, attempt, config_id, test_id, runner_id, job_name,
               caller_job, html_url, started_at, completed_at, duration_s,
               failed_step_name, failed_step_number, log_available,
               tested_revision, tested_artifact, collection_ref, collection_sha,
               signature_id, router_rule_id
        FROM jobs ORDER BY run_id, attempt, job_id
        """,
    ),
    DumpTable(
        "snap_provenance",
        ("channel", "arch", "revision", "built_sha", "observed_at"),
        """
        SELECT channel, arch, revision, built_sha, observed_at
        FROM snap_provenance ORDER BY channel, arch, revision
        """,
    ),
    DumpTable(
        "signature_classifications",
        (
            "signature_id",
            "substrate",
            "ruleset_version",
            "failure_class",
            "subclass",
            "rule_id",
            "confidence",
            "classified_by",
            "classified_at",
        ),
        """
        SELECT signature_id, substrate, ruleset_version, failure_class, subclass,
               rule_id, confidence, classified_by, classified_at
        FROM signature_classifications
        ORDER BY signature_id, substrate, ruleset_version, classified_by
        """,
    ),
    DumpTable(
        "retry_outcomes",
        ("run_id", "attempt", "config_id", "test_id", "conclusion"),
        """
        SELECT run_id, attempt, config_id, test_id, conclusion
        FROM retry_outcomes ORDER BY run_id, attempt, config_id, test_id
        """,
    ),
    DumpTable(
        "issues",
        (
            "claim_id",
            "signature_id",
            "substrate",
            "idempotency_key",
            "issue_number",
            "issue_state",
            "claimed_at",
            "opened_at",
            "closed_at",
            "closed_reason",
            "fixed_at_sha",
            "reopened_count",
            "post_attempts",
            "last_comment_at",
            "last_reported_occurrences",
            "occurrences_at_open",
            "opened_by_ruleset_version",
        ),
        """
        SELECT claim_id, signature_id, substrate, idempotency_key, issue_number,
               issue_state, claimed_at, opened_at, closed_at, closed_reason,
               fixed_at_sha, reopened_count, post_attempts, last_comment_at,
               last_reported_occurrences, occurrences_at_open,
               opened_by_ruleset_version
        FROM issues ORDER BY claim_id
        """,
    ),
)

DELETE_DURABLE_SQL = tuple(f"DELETE FROM {table}" for table in reversed(DURABLE_TABLES))


def _stable_id(namespace: str, values: Sequence[Any]) -> int:
    key = "\x1f".join("" if value is None else str(value) for value in values)
    digest = hashlib.sha256(f"{namespace}\x1f{key}".encode("utf-8")).digest()
    return int.from_bytes(digest[:8], "big") & ((1 << 63) - 1) or 1


def _label_key(labels: Iterable[str]) -> str:
    return ",".join(sorted(label for label in labels if label))


def _default_dump_path(db_path: Path) -> Path:
    return db_path.with_suffix(".sql")


def _atomic_write(path: Path, data: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_name(f".{path.name}.{os.getpid()}.tmp")
    try:
        tmp.write_bytes(data)
        tmp.replace(path)
    finally:
        if tmp.exists():
            tmp.unlink()


def initialise_database(conn: sqlite3.Connection) -> None:
    conn.execute("PRAGMA foreign_keys = ON")
    conn.executescript(SCHEMA_SQL)
    conn.execute(
        "INSERT OR REPLACE INTO schema_meta (key, value) VALUES (?, ?)",
        ("schema_version", str(SCHEMA_VERSION)),
    )
    conn.execute(
        "INSERT OR IGNORE INTO tests (test_id, nodeid, file, owner) VALUES (?, ?, ?, ?)",
        (0, "", None, None),
    )


def open_database(path: Path, dump_path: Optional[Path] = None) -> sqlite3.Connection:
    path.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(path)
    conn.execute("PRAGMA foreign_keys = ON")
    exists = conn.execute(
        "SELECT name FROM sqlite_master WHERE type = ? AND name = ?",
        ("table", "schema_meta"),
    ).fetchone()
    if exists is None:
        with conn:
            initialise_database(conn)
            source = dump_path or _default_dump_path(path)
            if source.exists():
                load_dump(conn, source)
    return conn


def rebuild_database(db_path: Path, dump_path: Path) -> sqlite3.Connection:
    if not dump_path.exists():
        raise FileNotFoundError(dump_path)
    if db_path.exists():
        db_path.unlink()
    conn = open_database(db_path, dump_path=dump_path)
    return conn


def load_dump(conn: sqlite3.Connection, dump_path: Path) -> None:
    conn.execute("PRAGMA foreign_keys = ON")
    for statement in DELETE_DURABLE_SQL:
        conn.execute(statement)
    conn.executescript(dump_path.read_text(encoding="utf-8"))
    conn.execute("PRAGMA foreign_keys = ON")
    row = conn.execute(
        "SELECT value FROM schema_meta WHERE key = ?", ("schema_version",)
    ).fetchone()
    if row != (str(SCHEMA_VERSION),):
        raise RuntimeError(
            f"metrics schema version mismatch: got {row[0] if row else 'missing'}, "
            f"expected {SCHEMA_VERSION}"
        )


def _quote(conn: sqlite3.Connection, value: Any) -> str:
    row = conn.execute("SELECT quote(?)", (value,)).fetchone()
    return str(row[0])


def dump_database_bytes(conn: sqlite3.Connection) -> bytes:
    lines = ["BEGIN TRANSACTION;"]
    for table in DUMP_TABLES:
        columns = ", ".join(table.columns)
        for row in conn.execute(table.select_sql):
            values = ", ".join(_quote(conn, value) for value in row)
            statement = f"INSERT INTO {table.name} ({columns}) VALUES ({values})"
            lines.append(statement + "\x3b")
    lines.append("COMMIT;")
    return ("\n".join(lines) + "\n").encode("utf-8")


def dump_database(conn: sqlite3.Connection, dump_path: Path) -> bytes:
    data = dump_database_bytes(conn)
    if dump_path.exists() and dump_path.read_bytes() == data:
        return data
    _atomic_write(dump_path, data)
    return data


def _get_or_create_workflow(conn: sqlite3.Connection, record: Any) -> int:
    workflow_id = _stable_id("workflow", (record.workflow_name,))
    conn.execute(
        """
        INSERT INTO workflows (workflow_id, name, file) VALUES (?, ?, ?)
        ON CONFLICT(name) DO UPDATE SET file = excluded.file
        """,
        (workflow_id, record.workflow_name, record.workflow_file),
    )
    row = conn.execute(
        "SELECT workflow_id FROM workflows WHERE name = ?", (record.workflow_name,)
    ).fetchone()
    return int(row[0])


def _get_or_create_config(
    conn: sqlite3.Connection, key: Tuple[str, str, str, str, str]
) -> int:
    config_id = _stable_id("config", key)
    conn.execute(
        """
        INSERT OR IGNORE INTO configs
        (config_id, os, arch, channel, flavor, substrate)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        (config_id, *key),
    )
    row = conn.execute(
        """
        SELECT config_id FROM configs
        WHERE os = ? AND arch = ? AND channel = ? AND flavor = ? AND substrate = ?
        """,
        key,
    ).fetchone()
    return int(row[0])


def _get_or_create_test(
    conn: sqlite3.Connection, nodeid: str, file: Optional[str]
) -> int:
    if not nodeid:
        return 0
    test_id = _stable_id("test", (nodeid,))
    conn.execute(
        """
        INSERT INTO tests (test_id, nodeid, file) VALUES (?, ?, ?)
        ON CONFLICT(nodeid) DO UPDATE SET file = excluded.file
        """,
        (test_id, nodeid, file),
    )
    row = conn.execute(
        "SELECT test_id FROM tests WHERE nodeid = ?", (nodeid,)
    ).fetchone()
    return int(row[0])


def _get_or_create_runner(conn: sqlite3.Connection, job: Any) -> Optional[int]:
    if not job.runner_labels:
        return None
    label_key = _label_key(job.runner_labels)
    runner_id = _stable_id("runner", (label_key,))
    conn.execute(
        """
        INSERT INTO runners (runner_id, label_key, self_hosted) VALUES (?, ?, ?)
        ON CONFLICT(label_key) DO UPDATE SET self_hosted = excluded.self_hosted
        """,
        (runner_id, label_key, int(job.self_hosted)),
    )
    row = conn.execute(
        "SELECT runner_id FROM runners WHERE label_key = ?", (label_key,)
    ).fetchone()
    runner_id = int(row[0])
    for label in job.runner_labels:
        conn.execute(
            "INSERT OR IGNORE INTO runner_labels (runner_id, label) VALUES (?, ?)",
            (runner_id, label),
        )
    return runner_id


def _upsert_run(conn: sqlite3.Connection, record: Any, workflow_id: int) -> None:
    conn.execute(
        """
        INSERT INTO runs (
            run_id, attempt, workflow_id, event, head_branch, head_sha, pr_number,
            created_at, started_at, completed_at, conclusion, jobs_total,
            jobs_success, jobs_failure, jobs_skipped, jobs_cancelled,
            runner_minutes_total, runner_minutes_failed, ingest_version
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(run_id, attempt) DO UPDATE SET
            workflow_id = excluded.workflow_id,
            event = excluded.event,
            head_branch = excluded.head_branch,
            head_sha = excluded.head_sha,
            pr_number = excluded.pr_number,
            created_at = excluded.created_at,
            started_at = excluded.started_at,
            completed_at = excluded.completed_at,
            conclusion = excluded.conclusion,
            jobs_total = excluded.jobs_total,
            jobs_success = excluded.jobs_success,
            jobs_failure = excluded.jobs_failure,
            jobs_skipped = excluded.jobs_skipped,
            jobs_cancelled = excluded.jobs_cancelled,
            runner_minutes_total = excluded.runner_minutes_total,
            runner_minutes_failed = excluded.runner_minutes_failed,
            ingest_version = excluded.ingest_version
        """,
        (
            record.run_id,
            record.attempt,
            workflow_id,
            record.event,
            record.head_branch,
            record.head_sha,
            record.pr_number,
            record.created_at,
            record.started_at,
            record.completed_at,
            record.conclusion,
            record.jobs_total,
            record.jobs_success,
            record.jobs_failure,
            record.jobs_skipped,
            record.jobs_cancelled,
            record.runner_minutes_total,
            record.runner_minutes_failed,
            record.ingest_version,
        ),
    )


def _upsert_job(
    conn: sqlite3.Connection,
    job: Any,
    config_id: int,
    test_id: int,
    runner_id: Optional[int],
) -> None:
    conn.execute(
        """
        INSERT INTO jobs (
            job_id, run_id, attempt, config_id, test_id, runner_id, job_name,
            caller_job, html_url, started_at, completed_at, duration_s,
            failed_step_name, failed_step_number, log_available,
            tested_revision, tested_artifact, collection_ref, collection_sha
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(job_id) DO UPDATE SET
            run_id = excluded.run_id,
            attempt = excluded.attempt,
            config_id = excluded.config_id,
            test_id = excluded.test_id,
            runner_id = excluded.runner_id,
            job_name = excluded.job_name,
            caller_job = excluded.caller_job,
            html_url = excluded.html_url,
            started_at = excluded.started_at,
            completed_at = excluded.completed_at,
            duration_s = excluded.duration_s,
            failed_step_name = excluded.failed_step_name,
            failed_step_number = excluded.failed_step_number,
            log_available = excluded.log_available,
            tested_revision = excluded.tested_revision,
            tested_artifact = excluded.tested_artifact,
            collection_ref = excluded.collection_ref,
            collection_sha = excluded.collection_sha
        """,
        (
            job.job_id,
            job.run_id,
            job.attempt,
            config_id,
            test_id,
            runner_id,
            job.job_name,
            job.caller_job,
            job.html_url,
            job.started_at,
            job.completed_at,
            job.duration_s,
            job.failed_step_name,
            job.failed_step_number,
            job.log_available,
            job.tested_revision,
            job.tested_artifact,
            job.collection_ref,
            job.collection_sha,
        ),
    )


def store_run_record(conn: sqlite3.Connection, record: Any) -> None:
    """Store one completed run attempt and its failed jobs idempotently."""

    if not record.conclusion:
        raise InFlightRunError(
            f"Refusing to store in-flight run {record.run_id}: conclusion is empty"
        )
    workflow_id = _get_or_create_workflow(conn, record)
    _upsert_run(conn, record, workflow_id)
    conn.execute(
        "DELETE FROM run_tests_collected WHERE run_id = ? AND attempt = ?",
        (record.run_id, record.attempt),
    )
    conn.execute(
        "DELETE FROM run_blocked_jobs WHERE run_id = ? AND attempt = ?",
        (record.run_id, record.attempt),
    )
    conn.execute(
        "DELETE FROM retry_outcomes WHERE run_id = ? AND attempt = ?",
        (record.run_id, record.attempt),
    )

    for key, count in record.tests_collected.items():
        config_id = _get_or_create_config(conn, key)
        conn.execute(
            """
            INSERT INTO run_tests_collected
            (run_id, attempt, config_id, count) VALUES (?, ?, ?, ?)
            ON CONFLICT(run_id, attempt, config_id) DO UPDATE SET
                count = excluded.count
            """,
            (record.run_id, record.attempt, config_id, count),
        )

    for prepare_job_name, jobs_not_run in record.blocked_jobs.items():
        conn.execute(
            """
            INSERT INTO run_blocked_jobs
            (run_id, attempt, prepare_job_name, jobs_not_run) VALUES (?, ?, ?, ?)
            ON CONFLICT(run_id, attempt, prepare_job_name) DO UPDATE SET
                jobs_not_run = excluded.jobs_not_run
            """,
            (record.run_id, record.attempt, prepare_job_name, jobs_not_run),
        )

    for job in record.failed_jobs:
        config_id = _get_or_create_config(conn, job.config_key)
        test_id = _get_or_create_test(conn, job.test_nodeid, job.test_file)
        runner_id = _get_or_create_runner(conn, job)
        _upsert_job(conn, job, config_id, test_id, runner_id)
        conn.execute("DELETE FROM job_steps WHERE job_id = ?", (job.job_id,))
        for step in job.steps:
            conn.execute(
                """
                INSERT INTO job_steps
                (job_id, number, name, conclusion, duration_s) VALUES (?, ?, ?, ?, ?)
                """,
                (
                    job.job_id,
                    step.number,
                    step.name,
                    step.conclusion,
                    step.duration_s,
                ),
            )

    for outcome in record.retry_outcomes:
        config_id = _get_or_create_config(conn, outcome.config_key)
        test_id = _get_or_create_test(conn, outcome.test_nodeid, outcome.test_file)
        conn.execute(
            """
            INSERT INTO retry_outcomes
            (run_id, attempt, config_id, test_id, conclusion) VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(run_id, attempt, config_id, test_id) DO UPDATE SET
                conclusion = excluded.conclusion
            """,
            (outcome.run_id, outcome.attempt, config_id, test_id, outcome.conclusion),
        )
