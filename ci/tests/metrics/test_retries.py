#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for offline retry annotation.

Flake detection has to be derivable from stored attempts alone. If it were
only obtainable at ingest time, a store written with ``--no-flake-detection``
could never be repaired -- the per-attempt skip check means a re-ingest walks
straight past it -- and recovering the signal would cost a full refetch of the
backfill.
"""

from metrics.ingest import annotate_retries_offline


def failure(job_name, **kwargs):
    base = {"job_name": job_name, "retried": False, "retry_outcome": None}
    base.update(kwargs)
    return base


def attempt(n, names, lost_upstream=0):
    return {
        "attempt": n,
        "run_id": 1,
        "failures": [failure(name) for name in names],
        "jobs_not_run_due_to_upstream": lost_upstream,
    }


class TestAnnotateRetriesOffline:
    def test_absent_from_later_attempt_counts_as_recovered(self):
        records = [attempt(1, ["a", "b"]), attempt(2, ["b"])]
        assert annotate_retries_offline(records) == 2
        first = {f["job_name"]: f for f in records[0]["failures"]}
        assert first["a"]["retry_outcome"] == "success"
        assert first["b"]["retry_outcome"] == "failure"

    def test_final_attempt_is_never_marked_retried(self):
        records = [attempt(1, ["a"]), attempt(2, ["a"])]
        annotate_retries_offline(records)
        assert records[1]["failures"][0]["retried"] is False

    def test_single_attempt_run_is_untouched(self):
        records = [attempt(1, ["a"])]
        assert annotate_retries_offline(records) == 0
        assert records[0]["failures"][0]["retried"] is False

    def test_upstream_loss_blocks_the_success_inference(self):
        """A job absent because the matrix was deleted did not pass.

        Scoring it as a flake would invent recoveries out of an infra
        failure, and inflate M4 exactly when CI is at its worst.
        """
        records = [attempt(1, ["a"]), attempt(2, [], lost_upstream=40)]
        assert annotate_retries_offline(records) == 0
        assert records[0]["failures"][0]["retried"] is False

    def test_existing_annotation_is_preserved(self):
        records = [attempt(1, ["a"]), attempt(2, [])]
        records[0]["failures"][0].update(retried=True, retry_outcome="failure")
        assert annotate_retries_offline(records) == 0
        assert records[0]["failures"][0]["retry_outcome"] == "failure"

    def test_out_of_order_records_are_handled(self):
        records = [attempt(3, ["a"]), attempt(1, ["a", "b"]), attempt(2, ["a"])]
        annotate_retries_offline(records)
        first = {f["job_name"]: f for f in records[1]["failures"]}
        assert first["b"]["retry_outcome"] == "success"
        assert first["a"]["retry_outcome"] == "failure"
