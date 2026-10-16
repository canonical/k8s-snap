#
# Copyright 2026 Canonical, Ltd.
#
"""Fetch logs for router-deferred failed jobs and attach their signatures."""

from __future__ import annotations

import dataclasses
import logging
import sqlite3
from datetime import datetime, timedelta, timezone
from typing import Dict, Iterable, List, Optional, Protocol, Tuple

from metrics.gh import GitHubError
from metrics.router import RouterJob, route
from metrics.signature import (
    EXTRACTOR_VERSION,
    NORMALISER_VERSION,
    Fingerprint,
    fingerprint,
)

LOG = logging.getLogger(__name__)

LOG_FINGERPRINTED = 1
LOG_MISSING = 0
LOG_GONE = 2
# A missing log is only evidence of a lost runner while the log should still
# exist; older jobs may simply have aged out of GitHub's log retention.
LOST_RUNNER_EVIDENCE_WINDOW = timedelta(days=7)


class LogClient(Protocol):
    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        """Return ``(status, log)``; ``log`` is ``None`` when GitHub has none."""


@dataclasses.dataclass(frozen=True)
class PendingJob(RouterJob):
    completed_at: Optional[str] = None


def pending_jobs(
    conn: sqlite3.Connection,
    run_ids: Optional[Iterable[int]] = None,
    limit: Optional[int] = None,
) -> List[PendingJob]:
    """Return stored failed jobs whose log has never been looked at.

    ``log_available IS NULL`` is the "not yet fetched" marker: a fetch sets it
    to 1 (fingerprinted), 0 (no log for a fresh job: runner-lost evidence) or
    2 (log gone, cause unknown), so re-ingesting never downloads a log twice.
    """

    sql = """
        SELECT j.job_id, j.failed_step_name, j.log_available, j.duration_s,
               cfg.channel, j.completed_at
        FROM jobs j
        JOIN configs cfg USING (config_id)
        WHERE j.log_available IS NULL
          AND j.signature_id IS NULL
          AND j.router_rule_id IS NULL
    """
    params: List[int] = []
    run_ids = list(run_ids or [])
    if run_ids:
        sql += f" AND j.run_id IN ({', '.join('?' for _ in run_ids)})"
        params.extend(run_ids)
    sql += " ORDER BY j.job_id"
    if limit is not None:
        sql += " LIMIT ?"
        params.append(limit)
    return [
        PendingJob(
            job_id=int(row[0]),
            failed_step_name=row[1],
            log_available=row[2],
            duration_s=row[3],
            channel=row[4] or "",
            completed_at=row[5],
        )
        for row in conn.execute(sql, params)
    ]


def stale_jobs(
    conn: sqlite3.Connection,
    run_ids: Optional[Iterable[int]] = None,
    limit: Optional[int] = None,
) -> List[PendingJob]:
    """Return fingerprinted jobs whose signature predates the current pipeline.

    Bumping ``EXTRACTOR_VERSION`` or ``NORMALISER_VERSION`` changes signature
    IDs; without re-fingerprinting, old and new jobs of one failure would be
    split across two signatures. Newest first, so the jobs most likely to still
    have a retained log go first.
    """

    sql = """
        SELECT j.job_id, j.failed_step_name, j.log_available, j.duration_s,
               cfg.channel, j.completed_at
        FROM jobs j
        JOIN configs cfg USING (config_id)
        JOIN signatures s ON s.signature_id = j.signature_id
        LEFT JOIN job_excerpts je ON je.job_id = j.job_id
        WHERE j.log_available = ?
          AND j.router_rule_id IS NULL
          AND (COALESCE(je.extractor_version, 0) < ? OR s.normaliser_version < ?)
    """
    params: List[int] = [LOG_FINGERPRINTED, EXTRACTOR_VERSION, NORMALISER_VERSION]
    run_ids = list(run_ids or [])
    if run_ids:
        sql += f" AND j.run_id IN ({', '.join('?' for _ in run_ids)})"
        params.extend(run_ids)
    sql += " ORDER BY j.job_id DESC"
    if limit is not None:
        sql += " LIMIT ?"
        params.append(limit)
    return [
        PendingJob(
            job_id=int(row[0]),
            failed_step_name=row[1],
            log_available=row[2],
            duration_s=row[3],
            channel=row[4] or "",
            completed_at=row[5],
        )
        for row in conn.execute(sql, params)
    ]


def _missing_log_state(status: int, completed_at: Optional[str], now: datetime) -> int:
    """Decide whether an absent log is runner-lost evidence or just gone."""

    if status != 404 or not completed_at:
        return LOG_GONE
    try:
        completed = datetime.fromisoformat(completed_at.replace("Z", "+00:00"))
    except ValueError:
        return LOG_GONE
    if now - completed > LOST_RUNNER_EVIDENCE_WINDOW:
        return LOG_GONE
    return LOG_MISSING


def store_fingerprint(conn: sqlite3.Connection, job_id: int, fp: Fingerprint) -> None:
    """Record ``fp`` against ``job_id``; the first job seen is the exemplar.

    An existing row with the same ID has the same normalised text, so a newer
    pipeline that reproduces it upgrades the row's versions; otherwise its
    jobs would look stale to ``stale_jobs`` forever.
    """

    conn.execute(
        """
        INSERT INTO signatures (
            signature_id, normalised_pattern, exemplar_excerpt, exemplar_job_id,
            normaliser_version, extractor_version
        ) VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(signature_id) DO UPDATE SET
            normaliser_version = MAX(normaliser_version, excluded.normaliser_version),
            extractor_version = MAX(extractor_version, excluded.extractor_version)
        """,
        (
            fp.signature_id,
            fp.normalised,
            fp.excerpt,
            job_id,
            fp.normaliser_version,
            fp.extractor_version,
        ),
    )
    conn.execute(
        """
        INSERT OR REPLACE INTO job_excerpts (job_id, excerpt, source, extractor_version)
        VALUES (?, ?, ?, ?)
        """,
        (job_id, fp.excerpt, fp.source, fp.extractor_version),
    )
    conn.execute(
        """
        UPDATE jobs SET log_available = ?, signature_id = ?
        WHERE job_id = ? AND router_rule_id IS NULL
        """,
        (LOG_FINGERPRINTED, fp.signature_id, job_id),
    )


def fetch_and_fingerprint(
    client: LogClient,
    conn: sqlite3.Connection,
    jobs: Iterable[PendingJob],
    now: Optional[datetime] = None,
) -> Dict[str, int]:
    """Fetch, scrub and fingerprint the log of every Stage A-deferred job.

    Jobs Stage A settles from metadata alone are skipped (router verdicts are
    never overwritten). Each job commits on its own so a transient API error
    leaves only that job pending for the next ingest or ``fetch-logs`` pass.
    """

    now = now or datetime.now(timezone.utc)
    summary = {
        "fingerprinted": 0,
        "reextracted": 0,
        "log_missing": 0,
        "log_gone": 0,
        "routed": 0,
        "errors": 0,
    }
    signatures = set()
    for job in jobs:
        if not route(job).defer:
            summary["routed"] += 1
            continue
        try:
            status, log = client.download_job_log(job.job_id)
        except GitHubError as exc:
            LOG.warning("Could not fetch log for job %s: %s", job.job_id, exc)
            summary["errors"] += 1
            continue
        with conn:
            if log is None:
                if job.log_available == LOG_FINGERPRINTED:
                    # Re-extraction found the log expired: keep the signature
                    # it already has rather than discard a known cause.
                    state = LOG_GONE
                else:
                    state = _missing_log_state(status, job.completed_at, now)
                conn.execute(
                    "UPDATE jobs SET log_available = ? WHERE job_id = ?",
                    (state, job.job_id),
                )
                summary["log_missing" if state == LOG_MISSING else "log_gone"] += 1
                continue
            fp = fingerprint(log)
            store_fingerprint(conn, job.job_id, fp)
        signatures.add(fp.signature_id)
        summary["fingerprinted"] += 1
        if job.log_available == LOG_FINGERPRINTED:
            summary["reextracted"] += 1
    summary["signatures"] = len(signatures)
    return summary
