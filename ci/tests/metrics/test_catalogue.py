#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the cross-period signature catalogue.

A rollup only sees its own period, so the ``first_seen`` it computes means
"first seen this month". Age derived from that resets at every month boundary
and understates worst for the signatures that have been failing longest --
measured on real data, the top signature had been failing for 88 days and the
September digest called it 10 days old. The catalogue is what makes age mean
what the report claims it means.

It is also the only long-term memory the system has: per-run records live in
artifacts with 90-day retention, so once they expire the catalogue is the sole
remaining evidence of when a signature first appeared. That makes idempotent,
order-independent merging a correctness requirement rather than a nicety.
"""

from metrics.rollup import apply_catalogue, merge_catalogue


def rollup(period, signatures):
    return {"period": period, "signatures": signatures}


def sig(sid="a" * 16, **kwargs):
    base = {
        "signature_id": sid,
        "occurrences": 1,
        "first_seen": "2026-06-15T00:00:00Z",
        "last_seen": "2026-06-20T00:00:00Z",
        "failure_class": "product.bug",
        "tests_affected": [],
    }
    base.update(kwargs)
    return base


class TestMergeCatalogue:
    def test_first_seen_survives_a_later_period(self):
        cat = merge_catalogue({}, rollup("2026-06", [sig()]))
        cat = merge_catalogue(
            cat,
            rollup("2026-09", [sig(first_seen="2026-09-01T00:00:00Z")]),
        )
        assert cat["a" * 16]["first_seen"] == "2026-06-15T00:00:00Z"

    def test_replaying_an_old_period_does_not_rewind_last_seen(self):
        """Order independence: rebuilds happen, and must not lose ground."""
        cat = merge_catalogue(
            {}, rollup("2026-09", [sig(last_seen="2026-09-30T00:00:00Z")])
        )
        cat = merge_catalogue(cat, rollup("2026-06", [sig()]))
        assert cat["a" * 16]["last_seen"] == "2026-09-30T00:00:00Z"
        assert cat["a" * 16]["first_seen"] == "2026-06-15T00:00:00Z"

    def test_occurrences_are_per_period_not_summed(self):
        """Re-running a period must correct its contribution, not double it."""
        cat = merge_catalogue({}, rollup("2026-06", [sig(occurrences=10)]))
        cat = merge_catalogue(cat, rollup("2026-07", [sig(occurrences=5)]))
        assert cat["a" * 16]["total_occurrences"] == 15
        cat = merge_catalogue(cat, rollup("2026-06", [sig(occurrences=10)]))
        assert cat["a" * 16]["total_occurrences"] == 15

    def test_a_corrected_period_lowers_the_total(self):
        cat = merge_catalogue({}, rollup("2026-06", [sig(occurrences=10)]))
        cat = merge_catalogue(cat, rollup("2026-06", [sig(occurrences=3)]))
        assert cat["a" * 16]["total_occurrences"] == 3

    def test_reclassification_updates_attribution(self):
        """A new rule must not leave stale ownership behind forever."""
        cat = merge_catalogue({}, rollup("2026-06", [sig(failure_class="unknown")]))
        cat = merge_catalogue(
            cat, rollup("2026-09", [sig(failure_class="infra.runner")])
        )
        assert cat["a" * 16]["failure_class"] == "infra.runner"

    def test_tests_accumulate_across_periods(self):
        cat = merge_catalogue({}, rollup("2026-06", [sig(tests_affected=["t1"])]))
        cat = merge_catalogue(
            cat, rollup("2026-07", [sig(tests_affected=["t1", "t2"])])
        )
        assert cat["a" * 16]["tests_affected"] == ["t1", "t2"]

    def test_signatures_without_an_id_are_ignored(self):
        assert merge_catalogue({}, rollup("2026-06", [{"occurrences": 1}])) == {}


class TestApplyCatalogue:
    def test_age_is_measured_from_the_catalogue_not_the_period(self):
        cat = merge_catalogue({}, rollup("2026-06", [sig()]))
        recent = rollup("2026-09", [sig(first_seen="2026-09-10T00:00:00Z")])
        apply_catalogue(recent, cat)
        assert recent["signatures"][0]["first_seen"] == "2026-06-15T00:00:00Z"
        assert recent["signatures"][0]["age_days"] > 60

    def test_unknown_signatures_are_left_alone(self):
        recent = rollup("2026-09", [sig(sid="b" * 16)])
        apply_catalogue(recent, {})
        assert recent["signatures"][0]["first_seen"] == "2026-06-15T00:00:00Z"
