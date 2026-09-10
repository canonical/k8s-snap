#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for job-name parsing and run summarisation.

Every job name asserted here was observed in the live API for
canonical/k8s-snap. The shapes vary in both the number and the position of
matrix parameters, which is why parsing is content-based rather than
positional.
"""

import pytest

from metrics.ingest import find_failed_step, parse_job_name, summarise_jobs
from metrics.models import StepRecord


class TestParseJobName:
    def test_nightly_integration_job(self):
        parsed = parse_job_name(
            "Integration (ubuntu:26.04, arm64, 1.35-classic/edge) "
            "/ tests/test_snap_services.py::test_snap_services"
        )
        assert parsed.os == "ubuntu:26.04"
        assert parsed.arch == "arm64"
        assert parsed.channel == "1.35-classic/edge"
        assert parsed.substrate == "lxd"
        assert parsed.test_file == "tests/test_snap_services.py"
        assert parsed.test_nodeid == "tests/test_snap_services.py::test_snap_services"
        assert parsed.caller_base == "Integration"

    def test_parametrised_test_id_is_preserved(self):
        parsed = parse_job_name(
            "Integration (ubuntu:24.04, arm64, 1.34-classic/edge) "
            "/ tests/test_node_availability_zone.py::test_node_availability_zone[etcd-True]"
        )
        assert parsed.test_nodeid.endswith("[etcd-True]")
        assert parsed.test_file == "tests/test_node_availability_zone.py"

    def test_pr_job_has_arch_only(self):
        # PR runs schedule one job per test *file* and carry no os/channel.
        parsed = parse_job_name("Integration (amd64) / tests/test_storage.py")
        assert parsed.arch == "amd64"
        assert parsed.os is None
        assert parsed.channel is None
        assert parsed.test_file == "tests/test_storage.py"
        assert parsed.test_nodeid is None

    def test_spread_datastore_omits_os(self):
        # Two parameters, and the *first* is the arch -- positional parsing
        # would misread this as an OS.
        parsed = parse_job_name(
            "Spread tests across different datastores (amd64, latest/edge) "
            "/ tests/test_smoke.py::test_smoke"
        )
        assert parsed.arch == "amd64"
        assert parsed.channel == "latest/edge"
        assert parsed.os is None

    def test_fips_series_name_implies_multipass(self):
        parsed = parse_job_name(
            "Integration tests on FIPS machine (jammy, amd64, latest/edge) "
            "/ tests/test_smoke.py::test_smoke"
        )
        assert parsed.os == "jammy"
        assert parsed.substrate == "multipass"
        assert parsed.arch == "amd64"

    def test_cncf_conformance_job(self):
        parsed = parse_job_name(
            "CNCF conformance test (ubuntu:22.04, amd64, 1.32-classic/stable) "
            "/ tests/test_cncf_conformance.py::test_cncf_conformance"
        )
        assert parsed.channel == "1.32-classic/stable"
        assert parsed.caller_base == "CNCF conformance test"

    def test_prepare_environment_is_flagged(self):
        parsed = parse_job_name(
            "Integration (ubuntu:22.04, amd64, 1.35-classic/edge) / Prepare Environment"
        )
        assert parsed.is_prepare is True
        assert parsed.test_file is None

    def test_build_snap_with_empty_matrix_param(self):
        # The empty `patch` matrix value renders as a double space.
        parsed = parse_job_name("Build k8s-snap  amd64 / Build snap")
        assert parsed.arch == "amd64"

    def test_non_matrix_job(self):
        parsed = parse_job_name("Get e2e test tags")
        assert parsed.caller_job is None
        assert parsed.inner_name == "Get e2e test tags"
        assert parsed.arch is None

    def test_docs_linkcheck_ref_is_not_mistaken_for_a_dimension(self):
        parsed = parse_job_name("Docs linkcheck (release-1.32) / Linkcheck")
        assert parsed.arch is None
        assert parsed.os is None
        assert parsed.channel is None
        assert parsed.unparsed_params == ["release-1.32"]

    def test_combo_key_is_stable(self):
        a = parse_job_name(
            "Integration (ubuntu:24.04, amd64, latest/edge) / tests/test_a.py::test_a"
        )
        b = parse_job_name(
            "Integration (ubuntu:24.04, amd64, latest/edge) / tests/test_b.py::test_b"
        )
        assert a.combo == b.combo


class TestFindFailedStep:
    def test_returns_first_failed_step(self):
        steps = [
            StepRecord(name="Set up job", number=1, conclusion="success"),
            StepRecord(name="Setup LXD", number=2, conclusion="failure"),
            StepRecord(name="Run test", number=3, conclusion="failure"),
        ]
        assert find_failed_step(steps).name == "Setup LXD"

    def test_lost_runner_has_no_failed_step(self):
        # Real shape of job 101894558721: setup green, everything from the test
        # step onward is null, job still concluded as a failure. A log-based
        # classifier cannot see this at all.
        steps = [
            StepRecord(name="Set up job", number=1, conclusion="success"),
            StepRecord(name="Setup LXD", number=6, conclusion="success"),
            StepRecord(
                name="Run test_skip_services_stop_on_remove", number=11, conclusion=None
            ),
            StepRecord(name="Prepare inspection reports", number=12, conclusion=None),
        ]
        assert find_failed_step(steps) is None


class TestSummariseJobs:
    def _job(self, name, conclusion):
        return {"name": name, "conclusion": conclusion}

    def test_counts_and_denominators(self):
        jobs = [
            self._job(
                "Integration (ubuntu:24.04, amd64, latest/edge) / Prepare Environment",
                "success",
            ),
            self._job(
                "Integration (ubuntu:24.04, amd64, latest/edge) / tests/test_a.py::test_a",
                "success",
            ),
            self._job(
                "Integration (ubuntu:24.04, amd64, latest/edge) / tests/test_b.py::test_b",
                "failure",
            ),
            self._job(
                "Integration (ubuntu:24.04, amd64, latest/edge) / tests/test_c.py::test_c",
                "skipped",
            ),
        ]
        counts, collected, failed_prepare = summarise_jobs(jobs)

        assert counts["success"] == 2
        assert counts["failure"] == 1
        assert counts["skipped"] == 1
        assert failed_prepare == []
        # Prepare jobs must not inflate the test denominator.
        assert sum(collected.values()) == 3

    def test_failed_prepare_is_captured(self):
        jobs = [
            self._job(
                "Integration (ubuntu:24.04, amd64, latest/edge) / Prepare Environment",
                "failure",
            ),
        ]
        counts, collected, failed_prepare = summarise_jobs(jobs)
        assert counts["failure"] == 1
        assert len(failed_prepare) == 1
        assert collected == {}


if __name__ == "__main__":
    raise SystemExit(pytest.main([__file__, "-v"]))
