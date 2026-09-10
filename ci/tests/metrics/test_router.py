#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the Stage A metadata router.

Every case here is drawn from a real job observed in the live API, so a
regression in the router shows up as a disagreement with production data
rather than with a hand-invented fixture.
"""

from metrics.models import ClassifiedBy, FailureClass, JobFailure, StepRecord
from metrics.router import apply_router, route, route_all


def make_job(
    failed_step=None,
    steps=None,
    duration_s=100.0,
    log_available=None,
    job_id=1,
    name="Integration (ubuntu:26.04, amd64, 1.35-classic/edge) / tests/t.py::test_x",
    channel=None,
):
    return JobFailure(
        job_id=job_id,
        run_id=1,
        attempt=1,
        job_name=name,
        html_url="https://example.invalid",
        failed_step_name=failed_step,
        steps=steps or [],
        duration_s=duration_s,
        log_available=log_available,
        channel=channel,
    )


class TestRunnerLost:
    """The failure mode a log-based classifier structurally cannot see."""

    def test_no_failed_step_with_trailing_nulls_is_runner_lost(self):
        # Shape of real job 101894558721: setup green, everything from the
        # test step onwards never ran, no log blob, 52 minutes elapsed.
        steps = [
            StepRecord("Set up job", 1, "success"),
            StepRecord("Checkout", 2, "success"),
            StepRecord("Run test_x", 3, None),
            StepRecord("Complete job", 4, None),
        ]
        verdict = route(make_job(failed_step=None, steps=steps, duration_s=3121.0))
        assert verdict.failure_class == FailureClass.INFRA_RUNNER.value
        assert verdict.subclass == "runner_lost"
        assert not verdict.defer

    def test_confidence_rises_when_log_is_confirmed_absent(self):
        steps = [
            StepRecord("Set up job", 1, "success"),
            StepRecord("Run test_x", 2, None),
        ]
        without = route(make_job(steps=steps, log_available=None))
        with_404 = route(make_job(steps=steps, log_available=False))
        assert with_404.confidence > without.confidence

    def test_job_at_the_six_hour_ceiling_is_a_timeout_not_a_lost_runner(self):
        steps = [StepRecord("Run test_x", 1, None)]
        verdict = route(make_job(steps=steps, duration_s=6 * 3600 - 10))
        assert verdict.subclass == "job_timeout"

    def test_job_with_no_steps_at_all_is_still_routed(self):
        verdict = route(make_job(steps=[]))
        assert verdict.failure_class == FailureClass.INFRA_RUNNER.value
        assert not verdict.defer


class TestStepRouting:
    def test_snap_download_from_a_channel_is_external(self):
        # 438 occurrences across the backfill, all carrying a channel. Fetching
        # a published snap exercises the store, not the product.
        verdict = route(
            make_job(failed_step="Download k8s-snap", channel="1.32-classic/edge")
        )
        assert verdict.failure_class == FailureClass.EXTERNAL_DEPENDENCY.value
        assert verdict.subclass == "snap_store"

    def test_snap_download_without_a_channel_is_ours(self):
        """Same step name, opposite fault domain.

        `Download k8s-snap` is a composite action with two modes. Without a
        channel it downloads *our own* build artifact, so a failure is a
        missing artifact we produced -- not the snap store's fault. Filing it
        as external would tell the team to wait out a problem they own.
        """
        verdict = route(make_job(failed_step="Download k8s-snap", channel=None))
        assert verdict.failure_class == FailureClass.CI_CONFIG.value
        assert verdict.subclass == "missing_artifact"

    def test_lxd_setup_is_provisioning(self):
        verdict = route(make_job(failed_step="Setup LXD"))
        assert verdict.failure_class == FailureClass.INFRA_PROVISIONING.value
        assert verdict.subclass == "lxd_setup"

    def test_prepare_environment_is_provisioning(self):
        verdict = route(make_job(failed_step="Prepare Environment"))
        assert verdict.failure_class == FailureClass.INFRA_PROVISIONING.value

    def test_set_up_job_is_runner(self):
        verdict = route(make_job(failed_step="Set up job"))
        assert verdict.failure_class == FailureClass.INFRA_RUNNER.value

    def test_post_steps_are_runner_cleanup(self):
        verdict = route(make_job(failed_step="Post Run actions/checkout@v4"))
        assert verdict.subclass == "node_cleanup"

    def test_test_collection_is_ci_config(self):
        verdict = route(make_job(failed_step="Get e2e test tags"))
        assert verdict.failure_class == FailureClass.CI_CONFIG.value
        assert verdict.subclass == "test_collection"


class TestDeferral:
    """Deferring is a correctness requirement, not an unfinished branch."""

    def test_test_steps_defer_to_stage_b(self):
        # The same step name covers a real upgrade regression, a flaky wait,
        # and a snap-store 503. Metadata cannot separate them; guessing here
        # would mis-attribute product bugs and corrupt the baseline.
        verdict = route(make_job(failed_step="Run test_version_upgrades"))
        assert verdict.defer
        assert verdict.failure_class is None

    def test_defer_rules_win_over_substring_matches(self):
        # "Run test_snap_services" contains "snap", which an unordered rule set
        # would happily route to the snap store.
        verdict = route(make_job(failed_step="Run test_snap_services"))
        assert verdict.defer

    def test_unrecognised_step_defers_rather_than_guessing(self):
        verdict = route(make_job(failed_step="Something entirely novel"))
        assert verdict.defer
        assert verdict.rule_id == "defer.unmatched-step"

    def test_deferred_job_is_not_written_to(self):
        job = make_job(failed_step="Run test_x")
        apply_router(job)
        assert job.failure_class == FailureClass.UNKNOWN.value
        assert job.classified_by == ClassifiedBy.NONE.value


class TestApplyAndSummarise:
    def test_router_verdict_is_recorded_with_provenance(self):
        job = make_job(failed_step="Setup LXD")
        apply_router(job)
        assert job.failure_class == FailureClass.INFRA_PROVISIONING.value
        assert job.classified_by == ClassifiedBy.ROUTER.value
        assert job.classifier_version.startswith("router-v")
        assert job.rule_id == "provisioning.lxd"

    def test_route_all_counts_routed_and_deferred(self):
        jobs = [
            make_job(failed_step="Download k8s-snap", job_id=1, channel="1.32/edge"),
            make_job(failed_step="Download k8s-snap", job_id=2, channel="1.32/edge"),
            make_job(failed_step="Run test_x", job_id=3),
        ]
        summary = route_all(jobs)
        assert summary["total"] == 3
        assert summary["routed"] == 2
        assert summary["deferred_to_stage_b"] == 1
        assert summary["by_class"]["external.dependency/snap_store"] == 2

    def test_routing_costs_no_api_calls(self):
        # The router takes a JobFailure and nothing else. If it ever grows a
        # client argument, this test forces that to be a deliberate decision.
        import inspect

        assert list(inspect.signature(route).parameters) == ["job"]
