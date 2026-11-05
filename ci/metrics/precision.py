#
# Copyright 2026 Canonical, Ltd.
#
"""Rule precision check: a rule must not fire on healthy jobs.

Adapted from OpenStack elastic-recheck's query guideline: a fingerprint that
also matches successful jobs describes a symptom, not a root cause, and more
than 10% hits on passing jobs means it is not specific enough. Here a rule's
``excerpt_pattern`` is tested against the error-shaped lines of successful
jobs that ran the same tests as the failures it classifies; those are the
lines a healthy job prints while retrying through transient errors.
"""

from __future__ import annotations

import dataclasses
import logging
import sqlite3
from typing import Dict, List, Optional, Protocol, Sequence, Set, Tuple

from metrics.gh import GitHubError
from metrics.ingest import parse_job_name
from metrics.rules import Rule, RulePack, _contexts
from metrics.scrub import scrub
from metrics.signature import error_lines

LOG = logging.getLogger(__name__)

DEFAULT_SAMPLE = 10
DEFAULT_RUNS = 5
DEFAULT_MAX_SUCCESS_HIT_RATE = 0.10
MAX_EXAMPLE_CHARS = 200


class PrecisionClient(Protocol):
    def list_run_jobs(self, run_id: int, attempt: int) -> List[Dict]:
        raise NotImplementedError

    def download_job_log(self, job_id: int) -> Tuple[int, Optional[str]]:
        raise NotImplementedError


@dataclasses.dataclass(frozen=True)
class RuleReport:
    rule_id: str
    signatures: int
    jobs: int
    success_checked: int
    success_hits: int
    verdict: str
    examples: Tuple[str, ...] = ()

    @property
    def hit_rate(self) -> Optional[float]:
        if not self.success_checked:
            return None
        return self.success_hits / self.success_checked

    def as_dict(self) -> Dict:
        return {**dataclasses.asdict(self), "hit_rate": self.hit_rate}


def _job_key(name: str, nodeid: str = "") -> str:
    """Identify what a job tested, so passing and failing runs of it pair up."""

    return nodeid or parse_job_name(name).test_nodeid or name


def rule_coverage(
    conn: sqlite3.Connection, pack: RulePack
) -> Dict[str, Tuple[int, int, Set[str]]]:
    """Return ``rule_id -> (signatures, jobs, job keys)`` for winning matches."""

    coverage: Dict[str, Tuple[int, int, Set[str]]] = {
        rule.id: (0, 0, set()) for rule in pack.rules
    }
    for context in _contexts(conn):
        rule = pack.match(context)
        if rule is None:
            continue
        rows = conn.execute(
            """
            SELECT j.job_name, COALESCE(t.nodeid, '')
            FROM jobs j
            JOIN configs cfg USING (config_id)
            JOIN tests t USING (test_id)
            WHERE j.signature_id = ? AND cfg.substrate = ?
              AND j.router_rule_id IS NULL
            """,
            (context.signature_id, context.substrate),
        ).fetchall()
        signatures, jobs, keys = coverage[rule.id]
        keys.update(_job_key(name, nodeid) for name, nodeid in rows)
        coverage[rule.id] = (signatures + 1, jobs + len(rows), keys)
    return coverage


def successful_jobs(
    conn: sqlite3.Connection, client: PrecisionClient, runs: int = DEFAULT_RUNS
) -> Dict[str, List[int]]:
    """Return ``job key -> successful job IDs`` from the newest stored runs."""

    by_key: Dict[str, List[int]] = {}
    rows = conn.execute(
        """
        SELECT run_id, attempt FROM runs
        WHERE jobs_success > 0
        ORDER BY completed_at DESC, run_id DESC
        LIMIT ?
        """,
        (runs,),
    ).fetchall()
    for run_id, attempt in rows:
        try:
            jobs = client.list_run_jobs(int(run_id), int(attempt))
        except GitHubError as exc:
            LOG.warning("Could not list jobs for run %s: %s", run_id, exc)
            continue
        for job in jobs:
            if job.get("conclusion") != "success":
                continue
            key = _job_key(str(job.get("name", "")))
            by_key.setdefault(key, []).append(int(job["id"]))
    return by_key


def _verdict(rule: Rule, checked: int, hits: int, max_rate: float) -> str:
    if rule.excerpt_pattern is None:
        return "pinned" if rule.signature_ids else "metadata-only"
    if rule.signature_ids:
        # An exact signature already fixes the excerpt; the pattern only
        # narrows it further, so green-job hits cannot widen the rule.
        return "pinned"
    if not checked:
        return "unchecked"
    return "imprecise" if hits / checked > max_rate else "ok"


def check_rules(
    conn: sqlite3.Connection,
    pack: RulePack,
    client: Optional[PrecisionClient],
    *,
    sample: int = DEFAULT_SAMPLE,
    runs: int = DEFAULT_RUNS,
    max_success_hit_rate: float = DEFAULT_MAX_SUCCESS_HIT_RATE,
) -> List[RuleReport]:
    """Report each rule's coverage and its hit rate on successful jobs.

    Without a ``client`` only coverage is reported. A rule that has matched
    nothing yet (a freshly proposed one) is checked against a general sample
    of successful jobs instead of the tests it classifies.
    """

    coverage = rule_coverage(conn, pack)
    healthy = successful_jobs(conn, client, runs) if client is not None else {}
    general = [job_id for job_ids in healthy.values() for job_id in job_ids]
    logs: Dict[int, Optional[Tuple[str, ...]]] = {}

    def lines_for(job_id: int) -> Optional[Tuple[str, ...]]:
        if client is None:
            return None
        if job_id not in logs:
            try:
                _status, log = client.download_job_log(job_id)
            except GitHubError as exc:
                LOG.warning("Could not fetch log for job %s: %s", job_id, exc)
                log = None
            logs[job_id] = error_lines(log) if log is not None else None
        return logs[job_id]

    reports = []
    for rule in pack.rules:
        signatures, jobs, keys = coverage[rule.id]
        checked = hits = 0
        examples: List[str] = []
        if rule.excerpt_pattern is not None and not rule.signature_ids:
            if keys:
                candidates = [
                    job for key in sorted(keys) for job in healthy.get(key, [])
                ]
            else:
                candidates = general
            for job_id in candidates[:sample]:
                lines = lines_for(job_id)
                if lines is None:
                    continue
                checked += 1
                matched = [line for line in lines if rule.excerpt_pattern.search(line)]
                if matched:
                    hits += 1
                    if len(examples) < 2:
                        examples.append(
                            scrub(matched[0])[:MAX_EXAMPLE_CHARS].replace("`", "'")
                        )
        reports.append(
            RuleReport(
                rule_id=rule.id,
                signatures=signatures,
                jobs=jobs,
                success_checked=checked,
                success_hits=hits,
                verdict=_verdict(rule, checked, hits, max_success_hit_rate),
                examples=tuple(examples),
            )
        )
    return reports


def render_markdown(reports: Sequence[RuleReport], max_rate: float) -> str:
    """Render a rule precision table for a job summary or terminal."""

    lines = [
        "## Rule precision",
        "",
        f"A rule is `imprecise` when its pattern hits more than {max_rate:.0%} of "
        "sampled successful jobs; `unused` rules matched nothing in the store.",
        "",
        "| Rule | Signatures | Jobs | Green hits | Verdict |",
        "| --- | ---: | ---: | ---: | --- |",
    ]
    for report in reports:
        green = (
            f"{report.success_hits}/{report.success_checked}"
            if report.success_checked
            else "-"
        )
        verdict = report.verdict
        if not report.jobs:
            verdict += ", unused"
        lines.append(
            f"| `{report.rule_id}` | {report.signatures} | {report.jobs} | "
            f"{green} | {verdict} |"
        )
    flagged = [report for report in reports if report.examples]
    if flagged:
        lines.extend(["", "### Green-job hits", ""])
        for report in flagged:
            for example in report.examples:
                lines.append(f"- `{report.rule_id}`: `{example}`")
    return "\n".join(lines) + "\n"
