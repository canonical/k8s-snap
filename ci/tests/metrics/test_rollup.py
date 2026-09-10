#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for rollup aggregation and report rendering.

The behaviours pinned here are the ones that make the numbers *trustworthy*
rather than merely present:

* rates return ``None`` rather than ``0.0`` when nothing was measured, so a
  missing denominator cannot masquerade as a good result;
* uninspected failures are excluded from attribution and from the unclassified
  rate, so neither can be improved by fetching fewer logs;
* rollups produced under different rule packs refuse to be compared;
* integrity warnings fire when coverage shrinks.
"""

import datetime as dt

import pytest

from metrics.report import render_markdown, render_mattermost
from metrics.rollup import aggregate, delta, top_signatures

NOW = dt.datetime(2026, 9, 15, tzinfo=dt.timezone.utc)


def _sig(rollup, signature_id):
    return next(s for s in rollup["signatures"] if s["signature_id"] == signature_id)


def failure(**kwargs):
    """A failure record that has been inspected unless told otherwise."""
    base = {
        "job_id": 1,
        "run_id": 100,
        "failure_class": "product.bug",
        "subclass": "upgrade_failure",
        "signature_id": "a" * 16,
        "classified_by": "rules",
        "rule_id": "product-x",
        "log_available": True,
        "test_nodeid": "tests/test_a.py::test_a",
        "os": "ubuntu:24.04",
        "arch": "amd64",
        "channel": "1.32-classic/stable",
        "started_at": "2026-09-10T01:00:00Z",
        "retried": False,
        "retry_outcome": None,
    }
    base.update(kwargs)
    return base


def run(**kwargs):
    base = {
        "run_id": 100,
        "attempt": 1,
        "workflow_slug": "nightly-test",
        "conclusion": "failure",
        "created_at": "2026-09-10T00:00:00Z",
        "jobs_total": 100,
        "jobs_failure": 1,
        "jobs_skipped": 0,
        "jobs_not_run_due_to_upstream": 0,
        "failed_prepare_jobs": [],
        "runner_minutes_total": 100.0,
        "runner_minutes_failed": 10.0,
        "self_hosted_minutes": 0.0,
        "self_hosted_minutes_failed": 0.0,
        "tests_collected_by_combo": {},
        "failures": [failure()],
    }
    base.update(kwargs)
    return base


class TestRates:
    def test_empty_input_yields_no_rates(self):
        rollup = aggregate([], "2026-09", now=NOW)
        assert rollup["runs"] == 0
        # None, not 0.0 -- "we measured nothing" must not read as "nothing
        # failed", which is the failure mode that makes a dashboard lie.
        assert rollup["m1_scheduled_green_rate"] is None
        assert rollup["m3_job_failure_rate"] is None
        assert rollup["m9_unclassified_rate"] is None

    def test_scheduled_and_pr_green_rates_are_separated(self):
        records = [
            run(run_id=1, workflow_slug="nightly-test", conclusion="failure"),
            run(run_id=2, workflow_slug="nightly-test", conclusion="success"),
            run(run_id=3, workflow_slug="lint_and_integration", conclusion="success"),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m1_scheduled_green_rate"] == 50.0
        assert rollup["m2_pr_first_pass_rate"] == 100.0

    def test_retry_success_counts_as_a_flake(self):
        """Flake evidence lives on superseded attempts, and only there.

        A rerun run is stored once per attempt. The final attempt's failures
        are by definition not retried, so M4 can only come from the earlier
        ones -- which is also why they must be ingested at all.
        """
        records = [
            run(
                attempt=1,
                failures=[
                    failure(retried=True, retry_outcome="success"),
                    failure(job_id=2, retried=True, retry_outcome="failure"),
                ],
            ),
            run(attempt=2, failures=[failure(job_id=2)]),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m4_flake_rate"] == 50.0

    def test_superseded_attempts_do_not_inflate_volume(self):
        """Rerunning a red run must not make CI look worse.

        Volume metrics count each run once, at its latest attempt; otherwise
        the act of retrying would add jobs, failures and runner minutes.
        """
        records = [
            run(
                attempt=1,
                failures=[
                    failure(retried=True, retry_outcome="success"),
                    failure(job_id=2, retried=True, retry_outcome="failure"),
                ],
            ),
            run(attempt=2, failures=[failure(job_id=2)]),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["runs"] == 1
        assert rollup["jobs_total"] == run()["jobs_total"]
        assert _sig(rollup, "a" * 16)["occurrences"] == 1
        assert _sig(rollup, "a" * 16)["retry_occurrences"] == 2

    def test_a_signature_only_seen_before_a_retry_stays_visible(self):
        """A failure that a rerun papered over is exactly what M4 is for.

        It must not vanish from the catalogue just because it is absent from
        the final attempt -- but it must not count as failure volume either.
        """
        records = [
            run(
                attempt=1,
                failures=[
                    failure(
                        signature_id="b" * 16, retried=True, retry_outcome="success"
                    )
                ],
            ),
            run(attempt=2, failures=[]),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        stat = _sig(rollup, "b" * 16)
        assert stat["occurrences"] == 0
        assert stat["recovered_on_retry"] == 1


class TestInspection:
    def test_uninspected_failures_are_not_counted_as_unclassified(self):
        """Never-fetched logs are absent data, not a classifier miss.

        Counting them as unclassified would make M9 a proxy for fetch
        coverage, and would let the rate be "improved" by fetching less.
        """
        records = [
            run(
                jobs_failure=2,
                failures=[
                    failure(),
                    failure(
                        job_id=2,
                        failure_class="unknown",
                        subclass=None,
                        signature_id=None,
                        classified_by="none",
                        rule_id=None,
                        log_available=None,
                    ),
                ],
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m9_inspected_failures"] == 1
        assert rollup["m9_unclassified_rate"] == 0.0
        assert rollup["m9_inspection_coverage"] == 50.0
        assert "unknown" not in rollup["m3_by_class"]

    def test_router_verdict_counts_as_inspected_without_logs(self):
        """Stage A never fetches a log, but it is still a real verdict."""
        records = [
            run(
                failures=[
                    failure(
                        failure_class="infra.provisioning",
                        subclass="lxd_setup",
                        signature_id=None,
                        classified_by="router",
                        rule_id=None,
                        log_available=None,
                    )
                ]
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m9_inspected_failures"] == 1
        assert rollup["m3_by_class"]["infra.provisioning"]["count"] == 1

    def test_inspected_but_unmatched_is_unclassified(self):
        records = [
            run(
                failures=[
                    failure(
                        failure_class="unknown",
                        subclass=None,
                        classified_by="none",
                        rule_id=None,
                        log_available=True,
                    )
                ]
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m9_unclassified_rate"] == 100.0


class TestSignatures:
    def test_occurrences_accumulate_across_runs(self):
        records = [run(run_id=1), run(run_id=2)]
        rollup = aggregate(records, "2026-09", now=NOW)
        stats = _sig(rollup, "a" * 16)
        assert stats["occurrences"] == 2

    def test_distinct_tests_and_configs_are_tracked(self):
        """Spread is the tell that separates shared setup from a real bug."""
        records = [
            run(
                failures=[
                    failure(test_nodeid="tests/test_a.py::x", arch="amd64"),
                    failure(job_id=2, test_nodeid="tests/test_b.py::y", arch="arm64"),
                ]
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        stats = _sig(rollup, "a" * 16)
        assert len(stats["tests_affected"]) == 2
        assert len(stats["configs_affected"]) == 2

    def test_concentration_is_over_fingerprinted_failures(self):
        records = [
            run(
                jobs_failure=3,
                failures=[
                    failure(),
                    failure(job_id=2),
                    failure(
                        job_id=3,
                        signature_id=None,
                        failure_class="infra.runner",
                        subclass="runner_lost",
                        classified_by="router",
                        log_available=False,
                    ),
                ],
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m5_fingerprinted_failures"] == 2
        assert rollup["m5_top5_concentration"] == 100.0

    def test_top_signatures_are_ranked_by_occurrence(self):
        records = [
            run(
                failures=[
                    failure(),
                    failure(job_id=2),
                    failure(job_id=3, signature_id="b" * 16),
                ]
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        ranked = top_signatures(rollup, limit=2)
        assert [s["signature_id"] for s in ranked] == ["a" * 16, "b" * 16]

    def test_unowned_signatures_are_counted(self):
        records = [
            run(
                failures=[
                    failure(
                        signature_id="c" * 16,
                        rule_id=None,
                        failure_class="unknown",
                        subclass=None,
                    )
                ]
            )
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m7_unowned_signatures"] == 1


class TestIntegrity:
    def test_upstream_losses_are_surfaced(self):
        """A prepare failure deletes a whole matrix -- metrics must not improve."""
        records = [run(jobs_not_run_due_to_upstream=250, failed_prepare_jobs=["j"])]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m11_jobs_not_run_due_to_upstream"] == 250
        assert rollup["m11_failed_prepare_jobs"] == 1
        text = render_mattermost(rollup)
        assert "250 job(s) never ran" in text

    def test_tests_collected_is_tracked_per_matrix_cell(self):
        records = [
            run(run_id=1, tests_collected_by_combo={"ubuntu:24.04/amd64": 100}),
            run(run_id=2, tests_collected_by_combo={"ubuntu:24.04/amd64": 80}),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        cell = rollup["m10_tests_collected"]["ubuntu:24.04/amd64"]
        assert cell["min"] == 80
        assert cell["max"] == 100

    def test_shrinking_coverage_is_reported(self):
        before = aggregate(
            [run(run_id=1, tests_collected_by_combo={"ubuntu:24.04/amd64": 100})],
            "2026-08",
            now=NOW,
        )
        after = aggregate(
            [run(run_id=2, tests_collected_by_combo={"ubuntu:24.04/amd64": 40})],
            "2026-09",
            now=NOW,
        )
        text = render_mattermost(after, before)
        assert "test count dropped" in text
        assert "100->40" in text


class TestDelta:
    def test_delta_is_the_difference_in_percentage_points(self):
        a = {"ruleset_version": 1, "m3_job_failure_rate": 10.0}
        b = {"ruleset_version": 1, "m3_job_failure_rate": 4.0}
        assert delta(b, a, "m3_job_failure_rate") == pytest.approx(-6.0)

    def test_delta_refuses_across_ruleset_versions(self):
        """Changing the rules changes the measurement, not the CI.

        Comparing across rule packs would report classifier churn as if it
        were a real improvement, which is exactly the sort of false progress
        this system exists to prevent.
        """
        a = {"ruleset_version": 1, "m3_job_failure_rate": 10.0}
        b = {"ruleset_version": 2, "m3_job_failure_rate": 4.0}
        assert delta(b, a, "m3_job_failure_rate") is None

    def test_delta_is_none_without_a_previous_period(self):
        b = {"ruleset_version": 1, "m3_job_failure_rate": 4.0}
        assert delta(b, None, "m3_job_failure_rate") is None


class TestRendering:
    def test_digest_leads_with_ownership(self):
        rollup = aggregate([run()], "2026-09", now=NOW)
        text = render_mattermost(rollup)
        owners = text.index("Who owns")
        assert owners < text.index("Top signatures")
        assert owners < text.index("Headline")
        assert "product team" in text

    def test_digest_flags_partial_inspection_coverage(self):
        records = [
            run(
                jobs_failure=10,
                failures=[failure()],
            )
        ]
        text = render_mattermost(aggregate(records, "2026-09", now=NOW))
        assert "inspection coverage" in text

    def test_digest_does_not_nag_when_coverage_is_complete(self):
        text = render_mattermost(aggregate([run()], "2026-09", now=NOW))
        assert "inspection coverage" not in text

    def test_direction_of_good_is_metric_specific(self):
        """A falling failure rate and a falling green rate are not both good."""
        before = aggregate(
            [run(run_id=1, jobs_total=100, jobs_failure=10)], "2026-08", now=NOW
        )
        after = aggregate(
            [run(run_id=2, jobs_total=100, jobs_failure=1)], "2026-09", now=NOW
        )
        text = render_mattermost(after, before)
        failure_line = [ln for ln in text.splitlines() if "failed job(s) of" in ln][0]
        assert ":small_green_triangle_down:" in failure_line

    def test_markdown_renders_a_trend_table(self):
        rollups = [
            aggregate([run(run_id=1)], "2026-08", now=NOW),
            aggregate([run(run_id=2)], "2026-09", now=NOW),
        ]
        text = render_markdown(rollups)
        assert "| 2026-08 |" in text
        assert "| 2026-09 |" in text
        assert "Integrity" in text

    def test_markdown_handles_no_data(self):
        assert "No rollups available" in render_markdown([])

    def test_digest_survives_an_empty_period(self):
        text = render_mattermost(aggregate([], "2026-09", now=NOW))
        assert "nothing to attribute" in text
        assert "n/a" in text


class TestRetryAccounting:
    def test_unannotated_superseded_failures_are_still_visible(self):
        """`--no-flake-detection` must lose the outcome, not the failure.

        Superseded failures are excluded from volume by design. If they were
        also skipped when the retry annotation is absent they would vanish
        entirely, and the gap would look like a quiet run rather than
        missing data.
        """
        records = [
            run(attempt=1, failures=[failure(retried=False, retry_outcome=None)]),
            run(attempt=2, failures=[]),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert rollup["m4_flake_rate"] is None
        assert _sig(rollup, "a" * 16)["retry_occurrences"] == 1

    def test_retry_only_signatures_stay_out_of_the_top_table(self):
        """A signature with no failure volume is not a top failure.

        It has zero occurrences by construction, so listing it would put an
        "x0" row in the table the team uses to pick what to fix next.
        """
        records = [
            run(
                attempt=1,
                failures=[
                    failure(
                        signature_id="b" * 16, retried=True, retry_outcome="success"
                    )
                ],
            ),
            run(attempt=2, failures=[failure()]),
        ]
        rollup = aggregate(records, "2026-09", now=NOW)
        assert [s["signature_id"] for s in top_signatures(rollup)] == ["a" * 16]
        assert _sig(rollup, "b" * 16)["tests_affected"] == ["tests/test_a.py::test_a"]
