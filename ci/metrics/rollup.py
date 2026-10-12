#
# Copyright 2026 Canonical, Ltd.
#
"""Derived CI failure metrics.

Rollups are deliberately not stored. Every value in this module is derived from
the durable fact tables and schema views at report time.
"""

from __future__ import annotations

import sqlite3
import statistics
from collections import Counter, defaultdict
from typing import Any, Dict, Iterable, List, Optional, Sequence, Tuple

from metrics.models import FailureClass

ROLLUP_VERSION = 1
SCHEDULED_WORKFLOW_FILES = ("nightly-test.yaml", "weekly-test.yaml")


def rate(numerator: float, denominator: float) -> Optional[float]:
    """Return a percentage, or ``None`` when the denominator is empty."""

    if denominator == 0:
        return None
    return round(100.0 * numerator / denominator, 1)


def delta(
    current: Dict[str, Any], previous: Optional[Dict[str, Any]], key: str
) -> Optional[float]:
    """Return percentage-point delta between two derived rollups."""

    if previous is None:
        return None
    before = previous.get(key)
    after = current.get(key)
    if before is None or after is None:
        return None
    return round(float(after) - float(before), 1)


def derive_rollups(conn: sqlite3.Connection) -> List[Dict[str, Any]]:
    """Derive one rollup per data month, in stable period order."""

    periods = [
        row[0]
        for row in conn.execute(
            """
            SELECT DISTINCT substr(COALESCE(started_at, completed_at), 1, 7) AS period
            FROM runs
            WHERE period IS NOT NULL AND period != ''
            ORDER BY period
            """
        )
    ]
    return [derive_rollup(conn, period) for period in periods]


def derive_rollup(
    conn: sqlite3.Connection, period: Optional[str] = None
) -> Dict[str, Any]:
    """Derive all report metrics for ``period`` or all stored data."""

    run_scope = _run_scope(conn, period)
    run_keys = [(row["run_id"], row["attempt"]) for row in run_scope]
    run_key_set = set(run_keys)
    latest_jobs = _jobs(conn, run_keys)
    flakes = _flakes(conn, run_key_set, period)
    signatures = _signature_rows(conn, run_keys)
    class_rows = _class_rows(latest_jobs)
    tests_collected = _tests_collected(conn, run_keys)
    blocked_jobs = _blocked_jobs(conn, run_keys)

    jobs_total = sum(int(row["jobs_total"]) for row in run_scope)
    jobs_failure = sum(int(row["jobs_failure"]) for row in run_scope)
    blocked_total = sum(row["jobs_not_run"] for row in blocked_jobs)
    failed_prepare_jobs = len(blocked_jobs)
    job_fact_count = len(latest_jobs)
    jobs_with_log_state = sum(
        1 for row in latest_jobs if row["log_available"] is not None
    )
    jobs_with_logs = sum(1 for row in latest_jobs if row["log_available"] == 1)
    unclassified = sum(
        1
        for row in latest_jobs
        if row["failure_class"] == FailureClass.UNKNOWN.value
        or row["classified_by"] == "none"
    )

    scheduled_total = 0
    scheduled_green = 0
    pr_total = 0
    pr_first_pass = 0
    runs_by_workflow: Dict[str, Dict[str, int]] = {}
    for row in run_scope:
        workflow_key = row["workflow_file"] or row["workflow_name"]
        workflow_stats = runs_by_workflow.setdefault(
            workflow_key, {"total": 0, "green": 0, "first_attempt_green": 0}
        )
        workflow_stats["total"] += 1
        if row["conclusion"] == "success":
            workflow_stats["green"] += 1
            if row["attempt"] == 1:
                workflow_stats["first_attempt_green"] += 1

        if _is_scheduled(row):
            scheduled_total += 1
            if row["conclusion"] == "success":
                scheduled_green += 1
        elif row["event"] == "pull_request":
            pr_total += 1
            if row["attempt"] == 1 and row["conclusion"] == "success":
                pr_first_pass += 1

    latest_completed_at = max(
        (row["completed_at"] for row in run_scope if row["completed_at"]), default=None
    )
    signature_hits = sum(row["occurrences"] for row in signatures)
    top5_hits = sum(row["occurrences"] for row in signatures[:5])
    signature_ages = [
        row["age_days"] for row in signatures if row["age_days"] is not None
    ]
    runner_minutes_total = round(
        sum(float(row["runner_minutes_total"]) for row in run_scope), 1
    )
    runner_minutes_failed = round(
        sum(float(row["runner_minutes_failed"]) for row in run_scope), 1
    )

    return {
        "period": period or "all",
        "data_through": latest_completed_at,
        "rollup_version": ROLLUP_VERSION,
        "runs": len(run_scope),
        "runs_by_workflow": dict(sorted(runs_by_workflow.items())),
        "jobs_total": jobs_total,
        "jobs_failure": jobs_failure,
        "jobs_skipped": sum(int(row["jobs_skipped"]) for row in run_scope),
        "jobs_cancelled": sum(int(row["jobs_cancelled"]) for row in run_scope),
        "m1_scheduled_green_rate": rate(scheduled_green, scheduled_total),
        "m1_scheduled_green_numerator": scheduled_green,
        "m1_scheduled_runs": scheduled_total,
        "m2_pr_first_pass_rate": rate(pr_first_pass, pr_total),
        "m2_pr_first_pass_numerator": pr_first_pass,
        "m2_pr_runs": pr_total,
        "m3_job_failure_rate": rate(jobs_failure, jobs_total),
        "m3_by_class": class_rows,
        "m4_flake_rate": rate(len(flakes), job_fact_count),
        "m4_flake_numerator": len(flakes),
        "m4_flake_denominator": job_fact_count,
        "m5_top5_concentration": rate(top5_hits, signature_hits),
        "m5_fingerprinted_failures": signature_hits,
        "m6_signature_age_days_p50": (
            round(statistics.median(signature_ages), 1) if signature_ages else None
        ),
        "m6_signature_age_days_max": max(signature_ages) if signature_ages else None,
        "m7_unowned_signatures": sum(
            1
            for row in signatures
            if row["failure_class"] == FailureClass.UNKNOWN.value
            or row["classified_by"] == "none"
        ),
        "m8_runner_minutes_total": runner_minutes_total,
        "m8_runner_minutes_failed": runner_minutes_failed,
        "m8_runner_minutes_wasted_share": rate(
            runner_minutes_failed, runner_minutes_total
        ),
        "m9_unclassified_rate": rate(unclassified, job_fact_count),
        "m9_unclassified_numerator": unclassified,
        "m9_failed_job_facts": job_fact_count,
        "m10_tests_collected": tests_collected,
        "m11_jobs_not_run_due_to_prepare": blocked_total,
        "m11_failed_prepare_jobs": failed_prepare_jobs,
        "m11_failure_pressure_rate": rate(
            jobs_failure + blocked_total, jobs_total + blocked_total
        ),
        "m11_failure_pressure_numerator": jobs_failure + blocked_total,
        "m11_failure_pressure_denominator": jobs_total + blocked_total,
        "m12_failed_job_fact_coverage": rate(job_fact_count, jobs_failure),
        "m12_failed_job_fact_numerator": job_fact_count,
        "m12_failed_job_fact_denominator": jobs_failure,
        "m12_log_state_coverage": rate(jobs_with_log_state, job_fact_count),
        "m12_log_available_rate": rate(jobs_with_logs, job_fact_count),
        "runner_minutes_total": runner_minutes_total,
        "runner_minutes_failed": runner_minutes_failed,
        "signatures": signatures,
    }


def signature_catalogue(conn: sqlite3.Connection) -> List[Dict[str, Any]]:
    """Return the deterministic signature catalogue derived from current facts."""

    rows = []
    for row in conn.execute(
        """
        SELECT s.signature_id,
               st.occurrences,
               st.runs_seen,
               st.tests_affected,
               st.configs_affected,
               st.first_seen,
               st.last_seen,
               COALESCE(vc.failure_class, 'unknown') AS failure_class,
               vc.subclass,
               vc.rule_id,
               vc.classified_by,
               s.state,
               s.exemplar_excerpt,
               j.html_url AS exemplar_job_url
        FROM signatures s
        JOIN v_signature_stats st USING (signature_id)
        LEFT JOIN v_signature_class vc
          ON vc.signature_id = s.signature_id AND vc.substrate = '*'
        LEFT JOIN jobs j ON j.job_id = s.exemplar_job_id
        ORDER BY st.occurrences DESC, s.signature_id
        """
    ):
        payload = dict(row)
        payload["contexts"] = _signature_contexts(conn, row["signature_id"])
        rows.append(payload)
    return rows


def top_signatures(rollup: Dict[str, Any], limit: int = 15) -> List[Dict[str, Any]]:
    """Return signatures in report order."""

    return list(rollup.get("signatures") or [])[:limit]


def _row_factory(conn: sqlite3.Connection) -> None:
    if conn.row_factory is not sqlite3.Row:
        conn.row_factory = sqlite3.Row


def _run_scope(conn: sqlite3.Connection, period: Optional[str]) -> List[sqlite3.Row]:
    _row_factory(conn)
    params: Tuple[Any, ...]
    period_clause = ""
    if period:
        period_clause = "AND substr(COALESCE(r.started_at, r.completed_at), 1, 7) = ?"
        params = (period,)
    else:
        params = ()
    return list(
        conn.execute(
            """
            WITH latest AS (
                SELECT run_id, MAX(attempt) AS attempt
                FROM runs
                GROUP BY run_id
            )
            SELECT r.run_id, r.attempt, r.event, r.started_at, r.completed_at,
                   r.conclusion, r.jobs_total, r.jobs_failure, r.jobs_skipped,
                   r.jobs_cancelled, r.runner_minutes_total,
                   r.runner_minutes_failed, w.name AS workflow_name,
                   COALESCE(w.file, '') AS workflow_file
            FROM runs r
            JOIN latest USING (run_id, attempt)
            JOIN workflows w USING (workflow_id)
            WHERE 1 = 1 {period_clause}
            ORDER BY COALESCE(r.started_at, r.completed_at), r.run_id, r.attempt
            """.format(
                period_clause=period_clause
            ),
            params,
        )
    )


def _jobs(
    conn: sqlite3.Connection, run_keys: Sequence[Tuple[int, int]]
) -> List[sqlite3.Row]:
    if not run_keys:
        return []
    values, params = _run_key_values(run_keys)
    return list(
        conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT vc.job_id, vc.run_id, vc.attempt, vc.signature_id,
                   vc.failure_class, vc.subclass, vc.classified_by,
                   j.log_available
            FROM v_job_class vc
            JOIN jobs j USING (job_id)
            JOIN scope USING (run_id, attempt)
            ORDER BY vc.job_id
            """.format(
                values=values
            ),
            params,
        )
    )


def _flakes(
    conn: sqlite3.Connection,
    latest_run_keys: set[Tuple[int, int]],
    period: Optional[str],
) -> List[sqlite3.Row]:
    params: Tuple[Any, ...]
    period_clause = ""
    if period:
        period_clause = "AND substr(COALESCE(r.started_at, r.completed_at), 1, 7) = ?"
        params = (period,)
    else:
        params = ()
    rows = list(
        conn.execute(
            """
            SELECT vf.signature_id, vf.run_id, vf.attempt, vf.config_id, vf.test_id
            FROM v_flakes vf
            JOIN runs r ON r.run_id = vf.run_id AND r.attempt = vf.attempt
            WHERE 1 = 1 {period_clause}
            ORDER BY vf.signature_id, vf.run_id, vf.attempt, vf.config_id, vf.test_id
            """.format(
                period_clause=period_clause
            ),
            params,
        )
    )
    return [
        row
        for row in rows
        if (row["run_id"], row["attempt"]) not in latest_run_keys
        or (row["run_id"], row["attempt"] + 1) in latest_run_keys
    ]


def _class_rows(jobs: Iterable[sqlite3.Row]) -> Dict[str, Dict[str, Any]]:
    counts = Counter(row["failure_class"] for row in jobs)
    total = sum(counts.values())
    return {
        name: {"count": count, "share": rate(count, total)}
        for name, count in sorted(counts.items(), key=lambda item: (-item[1], item[0]))
    }


def _tests_collected(
    conn: sqlite3.Connection, run_keys: Sequence[Tuple[int, int]]
) -> List[Dict[str, Any]]:
    if not run_keys:
        return []
    values, params = _run_key_values(run_keys)
    grouped: Dict[str, List[int]] = defaultdict(list)
    for row in conn.execute(
        """
        WITH scope(run_id, attempt) AS (VALUES {values})
        SELECT cfg.os, cfg.arch, cfg.channel, cfg.flavor, cfg.substrate, rtc.count
        FROM run_tests_collected rtc
        JOIN configs cfg USING (config_id)
        JOIN scope USING (run_id, attempt)
        ORDER BY cfg.os, cfg.arch, cfg.channel, cfg.flavor, cfg.substrate,
                 rtc.run_id, rtc.attempt
        """.format(
            values=values
        ),
        params,
    ):
        grouped[_config_label(row)].append(int(row["count"]))
    return [
        {
            "config": config,
            "min": min(counts),
            "median": int(statistics.median(counts)),
            "max": max(counts),
        }
        for config, counts in sorted(grouped.items())
    ]


def _blocked_jobs(
    conn: sqlite3.Connection, run_keys: Sequence[Tuple[int, int]]
) -> List[Dict[str, Any]]:
    if not run_keys:
        return []
    values, params = _run_key_values(run_keys)
    return [
        dict(row)
        for row in conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT rbj.run_id, rbj.attempt, rbj.prepare_job_name, rbj.jobs_not_run
            FROM run_blocked_jobs rbj
            JOIN scope USING (run_id, attempt)
            ORDER BY rbj.run_id, rbj.attempt, rbj.prepare_job_name
            """.format(
                values=values
            ),
            params,
        )
    ]


def _signature_rows(
    conn: sqlite3.Connection, run_keys: Sequence[Tuple[int, int]]
) -> List[Dict[str, Any]]:
    if not run_keys:
        return []
    values, params = _run_key_values(run_keys)
    rows = list(
        conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT j.signature_id,
                   COUNT(*) AS occurrences,
                   COUNT(DISTINCT j.run_id) AS runs_seen,
                   COUNT(DISTINCT j.test_id) AS tests_affected,
                   COUNT(DISTINCT j.config_id) AS configs_affected,
                   MIN(r.started_at) AS first_seen,
                   MAX(r.started_at) AS last_seen,
                   s.exemplar_excerpt,
                   ej.html_url AS exemplar_job_url
            FROM jobs j
            JOIN scope USING (run_id, attempt)
            JOIN runs r USING (run_id, attempt)
            JOIN signatures s USING (signature_id)
            LEFT JOIN jobs ej ON ej.job_id = s.exemplar_job_id
            WHERE j.signature_id IS NOT NULL
            GROUP BY j.signature_id
            ORDER BY occurrences DESC, j.signature_id
            """.format(
                values=values
            ),
            params,
        )
    )
    return [_signature_payload(conn, row, run_keys) for row in rows]


def _signature_payload(
    conn: sqlite3.Connection, row: sqlite3.Row, run_keys: Sequence[Tuple[int, int]]
) -> Dict[str, Any]:
    values, params = _run_key_values(run_keys)
    class_counts = list(
        conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT vc.failure_class, vc.subclass, vc.classified_by, COUNT(*) AS count
            FROM v_job_class vc
            JOIN scope USING (run_id, attempt)
            WHERE vc.signature_id = ?
            GROUP BY vc.failure_class, vc.subclass, vc.classified_by
            ORDER BY count DESC, vc.failure_class, vc.subclass, vc.classified_by
            """.format(
                values=values
            ),
            (*params, row["signature_id"]),
        )
    )
    top_class = class_counts[0] if class_counts else {}
    tests = [
        item[0]
        for item in conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT DISTINCT t.nodeid
            FROM jobs j
            JOIN scope USING (run_id, attempt)
            JOIN tests t USING (test_id)
            WHERE j.signature_id = ? AND t.nodeid != ''
            ORDER BY t.nodeid
            """.format(
                values=values
            ),
            (*params, row["signature_id"]),
        )
    ]
    configs = [
        _config_label(item)
        for item in conn.execute(
            """
            WITH scope(run_id, attempt) AS (VALUES {values})
            SELECT DISTINCT cfg.os, cfg.arch, cfg.channel, cfg.flavor, cfg.substrate
            FROM jobs j
            JOIN scope USING (run_id, attempt)
            JOIN configs cfg USING (config_id)
            WHERE j.signature_id = ?
            ORDER BY cfg.os, cfg.arch, cfg.channel, cfg.flavor, cfg.substrate
            """.format(
                values=values
            ),
            (*params, row["signature_id"]),
        )
    ]
    first_seen = row["first_seen"]
    last_seen = row["last_seen"]
    return {
        "signature_id": row["signature_id"],
        "occurrences": row["occurrences"],
        "runs_seen": row["runs_seen"],
        "tests_affected": tests,
        "configs_affected": configs,
        "tests_affected_count": row["tests_affected"],
        "configs_affected_count": row["configs_affected"],
        "first_seen": first_seen,
        "last_seen": last_seen,
        "age_days": _age_days(first_seen, last_seen),
        "failure_class": top_class["failure_class"] if top_class else "unknown",
        "subclass": top_class["subclass"] if top_class else None,
        "classified_by": top_class["classified_by"] if top_class else "none",
        "rule_id": _rule_for_signature(conn, row["signature_id"]),
        "exemplar_job_url": row["exemplar_job_url"],
        "exemplar_excerpt": row["exemplar_excerpt"],
    }


def _rule_for_signature(conn: sqlite3.Connection, signature_id: str) -> Optional[str]:
    row = conn.execute(
        """
        SELECT rule_id
        FROM v_signature_class
        WHERE signature_id = ?
        ORDER BY (substrate = '*') DESC, substrate
        LIMIT 1
        """,
        (signature_id,),
    ).fetchone()
    return row["rule_id"] if row else None


def _signature_contexts(
    conn: sqlite3.Connection, signature_id: str
) -> List[Dict[str, Any]]:
    return [
        dict(row)
        for row in conn.execute(
            """
            SELECT substrate, occurrences, runs_seen, first_seen, last_seen
            FROM v_signature_context_stats
            WHERE signature_id = ?
            ORDER BY substrate
            """,
            (signature_id,),
        )
    ]


def _run_key_values(run_keys: Sequence[Tuple[int, int]]) -> Tuple[str, Tuple[int, ...]]:
    values = ", ".join(["(?, ?)"] * len(run_keys))
    params: List[int] = []
    for run_id, attempt in run_keys:
        params.extend([run_id, attempt])
    return values, tuple(params)


def _is_scheduled(row: sqlite3.Row) -> bool:
    workflow_file = row["workflow_file"]
    workflow_name = row["workflow_name"]
    return workflow_file in SCHEDULED_WORKFLOW_FILES or workflow_name in {
        "nightly-test",
        "weekly-test",
        "Nightly",
        "Weekly",
    }


def _config_label(row: sqlite3.Row) -> str:
    parts = [
        row["os"] or "unknown-os",
        row["arch"] or "unknown-arch",
        row["channel"] or "artifact",
        row["flavor"] or "unknown-flavor",
        row["substrate"] or "unknown-substrate",
    ]
    return "/".join(parts)


def _age_days(first_seen: Optional[str], last_seen: Optional[str]) -> Optional[float]:
    if not first_seen or not last_seen:
        return None
    first = _date_prefix(first_seen)
    last = _date_prefix(last_seen)
    if first is None or last is None:
        return None
    return float((last - first).days)


def _date_prefix(value: str) -> Optional[Any]:
    import datetime as dt

    try:
        return dt.date.fromisoformat(value[:10])
    except ValueError:
        return None
