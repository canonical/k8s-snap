#
# Copyright 2026 Canonical, Ltd.
#
"""Rendering of rollups into human-readable reports.

Two formats:

* ``mattermost`` -- the daily/weekly digest for the existing channel.
* ``markdown`` -- the longer trend report committed alongside the rollups.

The digest is deliberately *not* a status dump. The measured problem with the
current alerting is not that it lacks detail -- it posts a 2,555-node emoji
tree every night -- but that it carries no signal: scheduled suites are red as
a steady state, nothing is attributed, and nobody owns anything, so the alert
is reliably ignored.

So the digest leads with attribution and ownership, and states the delta
against the previous period. Detail that nobody acts on belongs in a threaded
reply or in the committed markdown report, not at the top.

Integrity metrics are rendered whenever they are non-zero *or* when they have
worsened, so that "CI got better" can always be checked against "or did we
just stop running things".
"""

from typing import Any, Dict, List, Optional

from metrics.models import FailureClass
from metrics.rollup import delta, top_signatures

# Ownership is per fault domain, which is the point of axis A: the question a
# digest must answer is not "what broke" but "whose morning is this".
CLASS_OWNER_HINT = {
    FailureClass.INFRA_RUNNER.value: "CI/infra",
    FailureClass.INFRA_PROVISIONING.value: "CI/infra",
    FailureClass.EXTERNAL_DEPENDENCY.value: "external -- usually wait/retry",
    FailureClass.CI_CONFIG.value: "CI owners",
    FailureClass.PRODUCT_BUG.value: "product team",
    FailureClass.TEST_BUG.value: "test owners",
    FailureClass.UNKNOWN.value: "unowned -- needs triage",
}

_TREND_BETTER = ":small_green_triangle_down:"
_TREND_WORSE = ":small_red_triangle:"


def _pct(value: Optional[float]) -> str:
    return "n/a" if value is None else f"{value:.1f}%"


def _signed(value: Optional[float], lower_is_better: bool = True) -> str:
    """Render a delta with an explicit direction marker.

    Signs alone are ambiguous in a metric list where some things should go up
    (green rate) and others down (failure rate), so the direction of *good* is
    encoded per metric rather than left to the reader.
    """
    if value is None:
        return ""
    if value == 0:
        return "  (no change)"
    better = value < 0 if lower_is_better else value > 0
    marker = _TREND_BETTER if better else _TREND_WORSE
    return f"  ({value:+.1f} pp {marker})"


def _ranked_classes(rows: Dict[str, Any]) -> List[Any]:
    """Rank classes by count.

    The rollup is serialised with sorted keys, so the ``most_common`` ordering
    does not survive storage and has to be restored at render time.
    """
    return sorted(rows.items(), key=lambda kv: (-kv[1]["count"], kv[0]))


def render_mattermost(
    rollup: Dict[str, Any],
    previous: Optional[Dict[str, Any]] = None,
    period_label: str = "",
) -> str:
    """Render the digest posted to the CI channel."""
    label = period_label or rollup.get("period", "")
    lines: List[str] = []
    lines.append(f"#### CI failure report -- {label}")

    failures = rollup.get("jobs_failure") or 0
    jobs = rollup.get("jobs_total") or 0
    lines.append(
        f"{failures} failed job(s) of {jobs} across {rollup.get('runs', 0)} run(s) "
        f"-- {_pct(rollup.get('m3_job_failure_rate'))}"
        f"{_signed(delta(rollup, previous, 'm3_job_failure_rate'))}"
    )

    # Attribution first. Everything downstream is detail.
    lines.append("")
    inspected = rollup.get("m9_inspected_failures") or 0
    lines.append(f"**Who owns today's failures** ({inspected} inspected)")
    class_rows = rollup.get("m3_by_class") or {}
    if not class_rows:
        lines.append("- nothing to attribute")
    for name, payload in _ranked_classes(class_rows):
        owner = CLASS_OWNER_HINT.get(name, "?")
        lines.append(
            f"- `{name}` -- {payload['count']} "
            f"({_pct(payload['share'])} of inspected) -- {owner}"
        )

    lines.append("")
    lines.append("**Top signatures**")
    tops = top_signatures(rollup, limit=5)
    if not tops:
        lines.append("- none")
    for sig in tops:
        owner = "unowned" if not sig.get("rule_id") else sig["rule_id"]
        age = sig.get("age_days")
        age_text = f", first seen {age:.0f}d ago" if age else ""
        tests = sig.get("tests_affected") or []
        spread = f", {len(tests)} test(s)" if len(tests) > 1 else ""
        lines.append(
            f"- `{sig['signature_id']}` x{sig['occurrences']} "
            f"-- {sig.get('failure_class')}/{sig.get('subclass')} "
            f"({owner}{age_text}{spread})"
        )
        url = sig.get("exemplar_job_url")
        if url:
            lines.append(f"  [example job]({url})")

    concentration = rollup.get("m5_top5_concentration")
    if concentration is not None:
        lines.append("")
        lines.append(
            f"Top 5 signatures account for {_pct(concentration)} of "
            f"{rollup.get('m5_fingerprinted_failures', 0)} fingerprinted "
            "failures"
            f"{_signed(delta(rollup, previous, 'm5_top5_concentration'))}"
        )

    lines.append("")
    lines.append("**Headline**")
    lines.append(
        f"- scheduled green rate: {_pct(rollup.get('m1_scheduled_green_rate'))}"
        f"{_signed(delta(rollup, previous, 'm1_scheduled_green_rate'), False)}"
    )
    lines.append(
        f"- PR first-pass rate: {_pct(rollup.get('m2_pr_first_pass_rate'))}"
        f"{_signed(delta(rollup, previous, 'm2_pr_first_pass_rate'), False)}"
    )
    lines.append(
        f"- flake rate (recovered on retry): {_pct(rollup.get('m4_flake_rate'))}"
        f"{_signed(delta(rollup, previous, 'm4_flake_rate'))}"
    )

    flaky = _flaky_signatures(rollup)
    if flaky:
        lines.append("")
        lines.append("**Papered over by re-runs** (quarantine candidates)")
        for sig in flaky:
            tests = sig.get("tests_affected") or []
            label = tests[0] if tests else sig["signature_id"]
            lines.append(
                f"- `{label}` recovered on retry {sig['recovered_on_retry']}x "
                f"of {sig['retry_occurrences']} -- "
                f"{sig.get('failure_class')}/{sig.get('subclass')}"
            )

    integrity = _integrity_lines(rollup, previous)
    if integrity:
        lines.append("")
        lines.append("**Integrity checks**")
        lines.extend(integrity)

    unowned = rollup.get("m7_unowned_signatures") or 0
    if unowned:
        lines.append("")
        lines.append(
            f":warning: {unowned} signature(s) have no rule and no owner. "
            "Add them to `ci/failure_rules.yaml` or they will keep failing "
            "silently into `unknown`."
        )

    lines.append("")
    lines.append(
        f"_ruleset v{rollup.get('ruleset_version', '?')} · "
        f"rollup v{rollup.get('rollup_version', '?')}_"
    )
    return "\n".join(lines)


def _flaky_signatures(rollup, limit=5):
    """Signatures a re-run makes disappear.

    These are the ones the team currently pays for twice -- once in runner
    minutes, once in the habit of clicking re-run instead of filing a bug.
    A flake rate alone names nobody and changes nothing, so the digest has
    to say *which* signatures the re-runs are hiding.
    """
    flaky = [
        s
        for s in rollup.get("signatures") or []
        if (s.get("recovered_on_retry") or 0) > 0
    ]
    flaky.sort(key=lambda s: (-(s.get("recovered_on_retry") or 0), s["signature_id"]))
    return flaky[:limit]


def _integrity_lines(
    rollup: Dict[str, Any], previous: Optional[Dict[str, Any]]
) -> List[str]:
    """Integrity metrics, shown when non-zero or when they have worsened.

    Suppressing these while they are healthy keeps the digest readable;
    showing them the moment they move keeps "we improved" honest.
    """
    lines: List[str] = []

    unclassified = rollup.get("m9_unclassified_rate")
    if unclassified:
        lines.append(
            f"- unclassified: {_pct(unclassified)} of "
            f"{rollup.get('m9_inspected_failures', 0)} inspected failure(s)"
            f"{_signed(delta(rollup, previous, 'm9_unclassified_rate'))}"
        )

    # Coverage is reported separately from the unclassified rate because a low
    # unclassified rate over a small inspected sample says nothing.
    coverage = rollup.get("m9_inspection_coverage")
    if coverage is not None and coverage < 99:
        lines.append(
            f"- inspection coverage: {_pct(coverage)} of failures have been "
            "inspected -- the class split above covers that subset only"
        )

    lost = rollup.get("m11_jobs_not_run_due_to_upstream") or 0
    if lost:
        lines.append(
            "- :rotating_light: {} job(s) never ran because an upstream job "
            "failed. They are absent from the failure count above, so the "
            "numbers flatter us by that much.".format(lost)
        )

    shrunk = _shrinking_matrices(rollup, previous)
    if shrunk:
        lines.append(
            "- :rotating_light: test count dropped in {} matrix cell(s) "
            "({}) -- fewer tests collected, so a lower failure count may "
            "mean less coverage, not less breakage.".format(
                len(shrunk), ", ".join(shrunk[:3])
            )
        )

    quarantined = rollup.get("m12_quarantined_signatures") or 0
    if quarantined:
        lines.append(f"- {quarantined} signature(s) quarantined")

    return lines


def _shrinking_matrices(
    rollup: Dict[str, Any], previous: Optional[Dict[str, Any]]
) -> List[str]:
    """Matrix cells collecting fewer tests than last period.

    The cheapest way to make a failure count fall is to collect fewer tests.
    This is the check that catches it.
    """
    if not previous:
        return []
    before = previous.get("m10_tests_collected") or {}
    after = rollup.get("m10_tests_collected") or {}
    shrunk = []
    for combo, stats in sorted(after.items()):
        prior = before.get(combo)
        if prior and stats["median"] < prior["median"]:
            shrunk.append(f"{combo} {prior['median']}->{stats['median']}")
    return shrunk


def render_markdown(
    rollups: List[Dict[str, Any]], title: str = "CI failure trend"
) -> str:
    """Render a multi-period trend report for committing to the repo."""
    lines: List[str] = [f"# {title}", ""]
    if not rollups:
        lines.append("_No rollups available._")
        return "\n".join(lines)

    lines.append("## Headline metrics")
    lines.append("")
    lines.append(
        "| Period | Runs | Jobs | Failed | Fail % | Scheduled green | "
        "PR first-pass | Unclassified | Inspected | Top-5 conc. |"
    )
    lines.append("|---|---|---|---|---|---|---|---|---|---|")
    for rollup in rollups:
        lines.append(
            "| {period} | {runs} | {jobs} | {failed} | {rate} | {green} | "
            "{pr} | {unclassified} | {coverage} | {conc} |".format(
                period=rollup.get("period", "?"),
                runs=rollup.get("runs", 0),
                jobs=rollup.get("jobs_total", 0),
                failed=rollup.get("jobs_failure", 0),
                rate=_pct(rollup.get("m3_job_failure_rate")),
                green=_pct(rollup.get("m1_scheduled_green_rate")),
                pr=_pct(rollup.get("m2_pr_first_pass_rate")),
                unclassified=_pct(rollup.get("m9_unclassified_rate")),
                coverage=_pct(rollup.get("m9_inspection_coverage")),
                conc=_pct(rollup.get("m5_top5_concentration")),
            )
        )

    latest = rollups[-1]
    lines.append("")
    lines.append(f"## Failure attribution -- {latest.get('period')}")
    lines.append("")
    lines.append("| Class | Count | Share of inspected | Owner |")
    lines.append("|---|---|---|---|")
    for name, payload in _ranked_classes(latest.get("m3_by_class") or {}):
        lines.append(
            f"| `{name}` | {payload['count']} | {_pct(payload['share'])} | "
            f"{CLASS_OWNER_HINT.get(name, '?')} |"
        )

    lines.append("")
    lines.append("## Top signatures")
    lines.append("")
    lines.append("| Signature | Count | Class | Rule | Age (d) | Tests | Example |")
    lines.append("|---|---|---|---|---|---|---|")
    for sig in top_signatures(latest, limit=15):
        url = sig.get("exemplar_job_url") or ""
        link = f"[job]({url})" if url else ""
        age = sig.get("age_days")
        lines.append(
            f"| `{sig['signature_id']}` | {sig['occurrences']} | "
            f"{sig.get('failure_class')}/{sig.get('subclass')} | "
            f"{sig.get('rule_id') or '_unowned_'} | "
            f"{'' if age is None else f'{age:.0f}'} | "
            f"{len(sig.get('tests_affected') or [])} | {link} |"
        )

    lines.append("")
    lines.append("## Integrity")
    lines.append("")
    lines.append(
        "These guard against apparent improvement caused by running less, "
        "rather than by breaking less."
    )
    lines.append("")
    lines.append(f"- Unclassified rate: {_pct(latest.get('m9_unclassified_rate'))}")
    lines.append(
        f"- Jobs never run due to an upstream failure: "
        f"{latest.get('m11_jobs_not_run_due_to_upstream', 0)}"
    )
    lines.append(
        f"- Failed `Prepare Environment` jobs: "
        f"{latest.get('m11_failed_prepare_jobs', 0)}"
    )
    lines.append(
        f"- Quarantined signatures: {latest.get('m12_quarantined_signatures', 0)}"
    )
    lines.append("")
    lines.append(
        f"_Generated from rollup v{latest.get('rollup_version')} · "
        f"ruleset v{latest.get('ruleset_version')}. Rollups produced under "
        "different ruleset versions are not directly comparable._"
    )
    return "\n".join(lines)
