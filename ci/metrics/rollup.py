#
# Copyright 2026 Canonical, Ltd.
#
"""Aggregation of classified run records into periodic rollups.

A rollup is the durable artefact. Per-run records expire with their logs
after 90 days, but rollups are small enough to keep forever, so they are what
a trend claim ultimately rests on.

Three tiers of metric, following the design:

* **Tier 1 (headline)** -- M1 scheduled green rate, M2 PR first-pass rate,
  M3 job failure rate by class, M4 flake rate.
* **Tier 2 (diagnostic)** -- M5 top-5 signature concentration, M6 signature
  age, M7 unowned signatures, M8 runner-minutes wasted.
* **Tier 3 (integrity)** -- M9 unclassified rate, M10 tests collected per
  matrix cell, M11 jobs never run because an upstream job failed, M12
  quarantine count and age.

Tier 3 is the part that makes the other two falsifiable. Every headline
metric here can be improved by accident rather than by fixing anything:
disable a test and M3 improves, let the classifier rot and the infra share
drops, break `Prepare Environment` and an entire matrix disappears from the
denominator. M9-M11 exist to make those moves visible, so they are computed
unconditionally and reported even when they are boring.

Rollups are stamped with the versions of the code that produced them. A
rollup produced under a different ruleset is not silently comparable to one
produced under this ruleset, and pretending otherwise is how a team convinces
itself it has improved.
"""

import collections
import dataclasses
import datetime as dt
import statistics
from typing import Any, Dict, Iterable, List, Optional, Tuple

from metrics.models import ClassifiedBy, FailureClass

ROLLUP_VERSION = 1

# Suites that are supposed to be green on a schedule. PR workflows are
# measured separately (M2) because a red PR run is a normal, useful outcome
# whereas a red nightly is a broken promise.
SCHEDULED_WORKFLOWS = ("nightly-test", "weekly-test")


def _parse_ts(value: Optional[str]) -> Optional[dt.datetime]:
    if not value:
        return None
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None


def _day(value: Optional[str]) -> Optional[str]:
    parsed = _parse_ts(value)
    return parsed.strftime("%Y-%m-%d") if parsed else None


@dataclasses.dataclass
class SignatureStat:
    """Aggregate view of one signature over the rollup period."""

    signature_id: str
    occurrences: int = 0
    runs_seen: int = 0
    failure_class: str = FailureClass.UNKNOWN.value
    subclass: Optional[str] = None
    rule_id: Optional[str] = None
    owner: Optional[str] = None
    first_seen: Optional[str] = None
    last_seen: Optional[str] = None
    exemplar_job_url: Optional[str] = None
    exemplar_excerpt: Optional[str] = None
    tests_affected: List[str] = dataclasses.field(default_factory=list)
    configs_affected: List[str] = dataclasses.field(default_factory=list)
    #: Hits on superseded attempts. Kept apart from ``occurrences`` so that a
    #: signature which only ever fails on first attempts stays visible without
    #: inflating the failure volume that the rest of the report is built on.
    retry_occurrences: int = 0
    recovered_on_retry: int = 0

    def age_days(self, now: Optional[dt.datetime] = None) -> Optional[float]:
        first = _parse_ts(self.first_seen)
        if not first:
            return None
        now = now or dt.datetime.now(dt.timezone.utc)
        return round((now - first).total_seconds() / 86400.0, 1)

    def to_dict(self) -> Dict[str, Any]:
        payload = dataclasses.asdict(self)
        payload["age_days"] = self.age_days()
        return payload


def _config_key(failure: Dict[str, Any]) -> str:
    parts = [failure.get("os"), failure.get("arch"), failure.get("channel")]
    return "/".join(p for p in parts if p) or "unknown"


def aggregate(
    records: Iterable[Dict[str, Any]],
    period: str,
    now: Optional[dt.datetime] = None,
) -> Dict[str, Any]:
    """Reduce run records to a rollup for ``period`` (a label, e.g. 2026-09).

    ``records`` are the stored run dicts, already classified. Nothing here
    calls the API: a rollup must be reproducible from stored evidence alone,
    or the history it underpins is not auditable.
    """
    now = now or dt.datetime.now(dt.timezone.utc)

    # Reruns are stored one record per attempt, because retry evidence only
    # exists on the superseded ones. For every *volume* metric the run must
    # still count once, or rerunning a red run would inflate job counts,
    # runner minutes and failure totals -- i.e. the act of retrying would
    # make CI look worse. Superseded attempts contribute retry evidence only.
    materialised = list(records)
    latest_attempt: Dict[Any, int] = {}
    for entry in materialised:
        key = (entry.get("workflow_slug"), entry.get("run_id"))
        attempt = int(entry.get("attempt") or 1)
        if attempt > latest_attempt.get(key, 0):
            latest_attempt[key] = attempt

    runs = 0
    runs_by_workflow: Dict[str, Dict[str, int]] = collections.defaultdict(
        lambda: {"total": 0, "green": 0, "first_attempt_green": 0}
    )
    jobs_total = jobs_failure = jobs_skipped = 0
    jobs_lost_upstream = 0
    failed_prepare_jobs = 0
    by_class: Dict[str, int] = collections.Counter()
    by_subclass: Dict[str, int] = collections.Counter()
    by_rule: Dict[str, int] = collections.Counter()
    unclassified = 0
    inspected = 0
    retried = flaked = inconclusive_retries = 0
    runner_minutes_total = runner_minutes_failed = 0.0
    self_hosted_minutes = self_hosted_minutes_failed = 0.0
    tests_collected: Dict[str, List[int]] = collections.defaultdict(list)
    signatures: Dict[str, SignatureStat] = {}
    daily: Dict[str, Dict[str, int]] = collections.defaultdict(
        lambda: {"runs": 0, "jobs": 0, "failures": 0}
    )

    for record in materialised:
        superseded = int(record.get("attempt") or 1) != latest_attempt.get(
            (record.get("workflow_slug"), record.get("run_id")), 1
        )
        if superseded:
            # Retry evidence only. Occurrences are always tallied, even when
            # the `retried` annotation is missing (`--no-flake-detection`),
            # so that a superseded failure is never silently dropped from
            # both the volume and the retry view at once. M4 itself still
            # requires the annotation, since it needs the outcome.
            for failure in record.get("failures") or []:
                outcome = failure.get("retry_outcome")
                recovered = outcome == "success"
                if failure.get("retried"):
                    if outcome in ("success", "failure"):
                        # Only a retry that actually reached a verdict says
                        # anything about flakiness. A cancelled or unknown
                        # retry would otherwise sit in the denominator and
                        # push the flake rate *down*, making CI look steadier
                        # precisely because somebody cancelled a run.
                        retried += 1
                        if recovered:
                            flaked += 1
                    else:
                        inconclusive_retries += 1
                signature_id = failure.get("signature_id")
                if not signature_id:
                    continue
                stat = signatures.get(signature_id)
                if stat is None:
                    stat = SignatureStat(
                        signature_id=signature_id,
                        failure_class=failure.get("failure_class")
                        or FailureClass.UNKNOWN.value,
                        subclass=failure.get("subclass"),
                        rule_id=failure.get("rule_id"),
                        exemplar_job_url=failure.get("html_url"),
                        exemplar_excerpt=failure.get("evidence_excerpt"),
                    )
                    signatures[signature_id] = stat
                stat.retry_occurrences += 1
                if recovered:
                    stat.recovered_on_retry += 1
                # A retry-only signature still needs a human-readable label:
                # the flake list is meant to name a test to quarantine, not a
                # hash to go look up.
                test = failure.get("test_nodeid")
                if test and test not in stat.tests_affected:
                    stat.tests_affected.append(test)
                config = _config_key(failure)
                if config not in stat.configs_affected:
                    stat.configs_affected.append(config)
            continue

        runs += 1
        workflow = record.get("workflow_slug") or record.get("workflow_name") or "?"
        stats = runs_by_workflow[workflow]
        stats["total"] += 1
        if record.get("conclusion") == "success":
            stats["green"] += 1
            if int(record.get("attempt") or 1) == 1:
                stats["first_attempt_green"] += 1

        jobs_total += int(record.get("jobs_total") or 0)
        jobs_failure += int(record.get("jobs_failure") or 0)
        jobs_skipped += int(record.get("jobs_skipped") or 0)
        jobs_lost_upstream += int(record.get("jobs_not_run_due_to_upstream") or 0)
        failed_prepare_jobs += len(record.get("failed_prepare_jobs") or [])

        runner_minutes_total += float(record.get("runner_minutes_total") or 0.0)
        runner_minutes_failed += float(record.get("runner_minutes_failed") or 0.0)
        self_hosted_minutes += float(record.get("self_hosted_minutes") or 0.0)
        self_hosted_minutes_failed += float(
            record.get("self_hosted_minutes_failed") or 0.0
        )

        for combo, count in (record.get("tests_collected_by_combo") or {}).items():
            tests_collected[combo].append(int(count))

        day = _day(record.get("created_at"))
        if day:
            bucket = daily[day]
            bucket["runs"] += 1
            bucket["jobs"] += int(record.get("jobs_total") or 0)
            bucket["failures"] += int(record.get("jobs_failure") or 0)

        seen_this_run = set()
        for failure in record.get("failures") or []:
            failure_class = failure.get("failure_class") or FailureClass.UNKNOWN.value
            # A failure only counts towards the unclassified rate once we have
            # actually looked at it. A failure whose logs were never fetched is
            # *uninspected*, not unclassified -- conflating the two makes M9
            # track log-fetch coverage rather than classifier quality, and
            # would let the rate "improve" by fetching fewer logs.
            if (
                failure.get("classified_by") != ClassifiedBy.NONE.value
                or failure.get("log_available") is not None
            ):
                inspected += 1
                # Attribution is only tallied over inspected failures. Folding
                # uninspected ones in as `unknown` would drown the ownership
                # breakdown in jobs nobody has looked at yet, and the fix for
                # that is fetch coverage, not triage.
                by_class[failure_class] += 1
                by_subclass[f"{failure_class}/{failure.get('subclass')}"] += 1
                if failure.get("rule_id"):
                    by_rule[failure["rule_id"]] += 1
                if failure_class == FailureClass.UNKNOWN.value:
                    unclassified += 1

            signature_id = failure.get("signature_id")
            if not signature_id:
                continue
            stat = signatures.get(signature_id)
            if stat is None:
                stat = SignatureStat(
                    signature_id=signature_id,
                    failure_class=failure_class,
                    subclass=failure.get("subclass"),
                    rule_id=failure.get("rule_id"),
                    exemplar_job_url=failure.get("html_url"),
                    exemplar_excerpt=failure.get("evidence_excerpt"),
                )
                signatures[signature_id] = stat
            stat.occurrences += 1
            if signature_id not in seen_this_run:
                stat.runs_seen += 1
                seen_this_run.add(signature_id)

            started = failure.get("started_at") or record.get("created_at")
            if started:
                if stat.first_seen is None or started < stat.first_seen:
                    stat.first_seen = started
                if stat.last_seen is None or started > stat.last_seen:
                    stat.last_seen = started

            test = failure.get("test_nodeid")
            if test and test not in stat.tests_affected:
                stat.tests_affected.append(test)
            config = _config_key(failure)
            if config not in stat.configs_affected:
                stat.configs_affected.append(config)

    ranked = sorted(signatures.values(), key=lambda s: (-s.occurrences, s.signature_id))
    total_signature_hits = sum(s.occurrences for s in ranked)
    top5 = sum(s.occurrences for s in ranked[:5])

    scheduled = [
        stats
        for workflow, stats in runs_by_workflow.items()
        if workflow in SCHEDULED_WORKFLOWS
    ]
    scheduled_total = sum(s["total"] for s in scheduled)
    scheduled_green = sum(s["green"] for s in scheduled)

    pr = [
        stats
        for workflow, stats in runs_by_workflow.items()
        if workflow not in SCHEDULED_WORKFLOWS
    ]
    pr_total = sum(s["total"] for s in pr)
    pr_first_pass = sum(s["first_attempt_green"] for s in pr)

    ages = [s.age_days(now) for s in ranked]
    ages = [a for a in ages if a is not None]

    return {
        "period": period,
        "generated_at": now.isoformat(),
        "rollup_version": ROLLUP_VERSION,
        "runs": runs,
        "runs_by_workflow": {k: dict(v) for k, v in sorted(runs_by_workflow.items())},
        # -- Tier 1: headline ------------------------------------------
        "m1_scheduled_green_rate": _rate(scheduled_green, scheduled_total),
        "m2_pr_first_pass_rate": _rate(pr_first_pass, pr_total),
        "m3_job_failure_rate": _rate(jobs_failure, jobs_total),
        "m3_by_class": {
            k: {"count": v, "share": _rate(v, inspected)}
            for k, v in by_class.most_common()
        },
        "m3_by_subclass": dict(by_subclass.most_common()),
        "m4_flake_rate": _rate(flaked, retried),
        # -- Tier 2: diagnostic ----------------------------------------
        # Denominated over fingerprinted failures, not all failures: a
        # failure with no signature cannot be attributed to one, so including
        # it would understate concentration purely because of fetch coverage.
        "m5_top5_concentration": _rate(top5, total_signature_hits),
        "m5_fingerprinted_failures": total_signature_hits,
        "m6_signature_age_days_p50": (
            round(statistics.median(ages), 1) if ages else None
        ),
        "m6_signature_age_days_max": max(ages) if ages else None,
        "m7_unowned_signatures": sum(1 for s in ranked if not s.rule_id),
        "m8_runner_minutes_total": round(runner_minutes_total, 1),
        "m8_runner_minutes_failed": round(runner_minutes_failed, 1),
        "m8_runner_minutes_wasted_share": _rate(
            runner_minutes_failed, runner_minutes_total
        ),
        "m8_self_hosted_minutes_failed": round(self_hosted_minutes_failed, 1),
        # -- Tier 3: integrity -----------------------------------------
        # Reported even when boring. Without these, every metric above can
        # be improved by deleting coverage rather than fixing anything.
        "m9_unclassified_rate": _rate(unclassified, inspected),
        "m9_inspected_failures": inspected,
        "m9_inspection_coverage": _rate(inspected, jobs_failure),
        "m10_tests_collected": {
            combo: {
                "min": min(counts),
                "max": max(counts),
                "median": int(statistics.median(counts)),
            }
            for combo, counts in sorted(tests_collected.items())
        },
        "m11_jobs_not_run_due_to_upstream": jobs_lost_upstream,
        "m11_failed_prepare_jobs": failed_prepare_jobs,
        "m12_quarantined_signatures": 0,
        "m12_quarantine_age_days_p50": None,
        # -- Supporting detail -----------------------------------------
        "jobs_total": jobs_total,
        "jobs_failure": jobs_failure,
        "jobs_skipped": jobs_skipped,
        "retried_jobs": retried,
        "inconclusive_retries": inconclusive_retries,
        "recovered_on_retry": flaked,
        "by_rule": dict(by_rule.most_common()),
        "signatures": [s.to_dict() for s in ranked],
        "daily": {k: dict(v) for k, v in sorted(daily.items())},
    }


def _rate(numerator: float, denominator: float) -> Optional[float]:
    """Rate as a percentage, or ``None`` when there is nothing to divide by.

    Returning ``None`` rather than 0.0 matters: "no PR runs this period" and
    "every PR run failed" are different facts, and rendering both as 0% would
    invent a trend out of missing data.
    """
    if not denominator:
        return None
    return round(100.0 * numerator / denominator, 1)


def delta(
    current: Dict[str, Any], previous: Optional[Dict[str, Any]], key: str
) -> Optional[float]:
    """Change in a metric against the previous rollup, or ``None``.

    Only comparable when both rollups were produced by the same ruleset;
    otherwise a change in the rules would masquerade as a change in CI.
    """
    if not previous:
        return None
    if previous.get("ruleset_version") != current.get("ruleset_version"):
        return None
    before, after = previous.get(key), current.get(key)
    if before is None or after is None:
        return None
    return round(after - before, 1)


def top_signatures(rollup: Dict[str, Any], limit: int = 5) -> List[Dict[str, Any]]:
    """The signatures carrying the most failure volume.

    Retry-only signatures (seen solely on a superseded attempt) are excluded:
    they have zero occurrences by construction, and listing them here as
    "x0" would put entries in the top-failures table that contributed no
    failures. They are reported in the flake section instead.
    """
    ranked = [
        s for s in (rollup.get("signatures") or []) if (s.get("occurrences") or 0) > 0
    ]
    return ranked[:limit]


def class_split(rollup: Dict[str, Any]) -> List[Tuple[str, int, Optional[float]]]:
    """(class, count, share) ordered by count, for reporting."""
    return [
        (name, payload["count"], payload["share"])
        for name, payload in (rollup.get("m3_by_class") or {}).items()
    ]
