#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for excerpt extraction, normalisation and signature hashing.

The fixtures below are reduced from real job logs of nightly run
34172139128. The reduction keeps the structural features that matter --
the exception chain, the timestamp prefixes, the runner paths -- because
those are precisely what the code under test reasons about.
"""

from metrics.signature import (
    MAX_EXCERPT_BYTES,
    extract,
    fingerprint,
    normalise,
    signature_id,
    strip_timestamps,
)


def as_log(body, stamp="2026-09-08T02:25:17.1935006Z"):
    """Prefix each line with a GitHub log timestamp, as the real API does.

    Applied at runtime rather than baked into the fixtures so the fixtures
    stay readable (and inside the line-length limit) while still exercising
    the timestamp-stripping path.
    """
    return "\n".join(f"{stamp} {line}" for line in body.strip("\n").splitlines())


UBUNTU_WORKSPACE = "/home/ubuntu/actions-runner/_work/k8s-snap/k8s-snap"
HOSTED_WORKSPACE = "/home/runner/work/k8s-snap/k8s-snap"

# Real shape: pytest wraps the wait in tenacity, so the *outer* exception is
# RetryError and the *inner* one is the actual bug.
CHAINED_LOG = as_log(
    """
=================================== FAILURES ===================================
____________________ test_version_downgrades_with_rollback _____________________
Traceback (most recent call last):
  File "{ws}/tests/util.py", line 362, in wait_until_k8s_ready
    check_snap_services_ready(instance)
AssertionError: Service kube-proxy should be active, but it is inactive

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "{ws}/tests/test_version_upgrades.py", line 231, in test_vu
    util.wait_until_k8s_ready(cp, instances)
tenacity.RetryError: RetryError[<Future at 0x736ed8a7faa0 state=finished raised>]
=========================== short test summary info ============================
FAILED tests/test_version_upgrades.py::test_vu - tenacity.RetryError
================= 1 failed, 60 warnings in 1733.92s (0:28:53) ==================
##[error]Process completed with exit code 1.
""".format(
        ws=UBUNTU_WORKSPACE
    )
)

# Same failure, different run: GitHub-hosted runner path, different container
# suffix, different addresses, different timings.
CHAINED_LOG_OTHER_RUN = as_log(
    """
=================================== FAILURES ===================================
____________________ test_version_downgrades_with_rollback _____________________
Traceback (most recent call last):
  File "{ws}/tests/util.py", line 371, in wait_until_k8s_ready
    check_snap_services_ready(instance)
AssertionError: Service kube-proxy should be active, but it is inactive

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "{ws}/tests/test_version_upgrades.py", line 240, in test_vu
    util.wait_until_k8s_ready(cp, instances)
tenacity.RetryError: RetryError[<Future at 0x7ffabc001122 state=finished raised>]
##[error]Process completed with exit code 1.
""".format(
        ws=HOSTED_WORKSPACE
    ),
    stamp="2026-09-15T04:11:09.9100000Z",
)

DIFFERENT_FAILURE_LOG = as_log(
    """
=================================== FAILURES ===================================
____________________________ test_snap_services ________________________________
Traceback (most recent call last):
  File "{ws}/tests/test_storage.py", line 35, in test_storage
    assert pvc_is_bound(claim)
AssertionError: PersistentVolumeClaim never reached Bound
##[error]Process completed with exit code 1.
""".format(
        ws=HOSTED_WORKSPACE
    )
)

NO_PYTEST_LOG = as_log(
    """
Setting up LXD
##[error]Failed to create storage pool: device busy
##[error]Process completed with exit code 1.
"""
)


class TestTimestampStripping:
    def test_github_prefix_is_removed(self):
        assert strip_timestamps(["2026-09-08T02:25:14.0000000Z hello"]) == ["hello"]

    def test_lines_without_a_prefix_are_untouched(self):
        assert strip_timestamps(["plain line"]) == ["plain line"]


class TestExtraction:
    def test_prefers_the_failures_block(self):
        assert extract(CHAINED_LOG).source == "failures-block"

    def test_extracts_the_root_cause_not_the_wrapper(self):
        # This is the whole point of the module. tenacity.RetryError is what
        # the pytest summary line reports, and it is identical for every
        # retry-wrapped failure in the suite -- useless as a fingerprint.
        text = extract(CHAINED_LOG).text
        assert "kube-proxy should be active" in text
        assert "tenacity.RetryError" not in text

    def test_keeps_the_calling_frame_for_discrimination(self):
        # Without a frame, two unrelated "AssertionError" failures would hash
        # identically.
        text = extract(CHAINED_LOG).text
        assert "wait_until_k8s_ready" in text

    def test_falls_back_to_summary_line_without_a_failures_block(self):
        log = as_log(
            """
=========================== short test summary info ============================
FAILED tests/test_x.py::test_y - ValueError: boom
"""
        )
        excerpt = extract(log)
        assert excerpt.source == "summary-line"
        assert "test_y" in excerpt.text

    def test_falls_back_to_error_marker_for_non_pytest_failures(self):
        excerpt = extract(NO_PYTEST_LOG)
        assert excerpt.source == "error-marker"
        assert "storage pool" in excerpt.text

    def test_generic_exit_code_error_is_not_treated_as_informative(self):
        # "Process completed with exit code 1" is what the annotations API
        # returns for every failed job; treating it as evidence would give
        # every failure the same signature.
        log = (
            "2026-09-08T01:00:00.0000000Z ##[error]Process completed with exit code 1."
        )
        assert extract(log).source != "error-marker"

    def test_empty_log_does_not_raise(self):
        assert extract("").text == ""


class TestNormalisation:
    def test_collapses_runner_paths(self):
        out = normalise(
            'File "/home/runner/work/k8s-snap/k8s-snap/tests/t.py", line 5, in f'
        )
        assert "/home/runner" not in out
        assert "tests/t.py" in out, "the interesting part of the path must survive"

    def test_self_hosted_and_hosted_paths_collapse_identically(self):
        hosted = normalise('File "/home/runner/work/k8s-snap/k8s-snap/tests/t.py"')
        self_hosted = normalise(
            'File "/home/ubuntu/actions-runner/_work/k8s-snap/k8s-snap/tests/t.py"'
        )
        assert hosted == self_hosted

    def test_collapses_line_numbers(self):
        assert normalise("line 362") == normalise("line 371")

    def test_collapses_container_names(self):
        assert normalise("lxc rm k8s-integration-1-9e535e-registry") == normalise(
            "lxc rm k8s-integration-2-ab12cd-registry"
        )

    def test_collapses_addresses_and_timings(self):
        assert normalise("at 0x736ed8a7faa0 after 1733.92s") == normalise(
            "at 0x7ffabc001122 after 12.10s"
        )

    def test_drops_caret_lines(self):
        # Caret width tracks identifier length, so hashing it would make the
        # signature depend on variable names.
        assert "^" not in normalise("assert f(x)\n    ^^^^^^^^\nAssertionError")

    def test_does_not_collapse_small_distinguishing_numbers(self):
        # Blanket digit stripping would merge genuinely different failures.
        assert normalise("exit code 1") != normalise("exit code 2")

    def test_preserves_the_error_message(self):
        out = normalise("AssertionError: Service kube-proxy should be active")
        assert "kube-proxy should be active" in out


class TestSignatureStability:
    """The property the entire measurement baseline rests on."""

    def test_same_failure_across_runs_hashes_identically(self):
        # Different day, different runner type, different paths, different
        # line numbers, different object addresses -- same bug.
        assert fingerprint(CHAINED_LOG).signature_id == (
            fingerprint(CHAINED_LOG_OTHER_RUN).signature_id
        )

    def test_different_failures_hash_differently(self):
        assert fingerprint(CHAINED_LOG).signature_id != (
            fingerprint(DIFFERENT_FAILURE_LOG).signature_id
        )

    def test_signature_is_deterministic(self):
        assert (
            fingerprint(CHAINED_LOG).signature_id
            == fingerprint(CHAINED_LOG).signature_id
        )

    def test_signature_id_is_short_and_hex(self):
        sig = signature_id("anything")
        assert len(sig) == 16
        int(sig, 16)


class TestFingerprintSafety:
    def test_excerpt_is_scrubbed(self):
        log = CHAINED_LOG.replace(
            "check_snap_services_ready(instance)",
            "connect(token='ghp_" + "a" * 36 + "')",
        )
        fp = fingerprint(log)
        assert "ghp_" not in fp.excerpt
        assert "ghp_" not in fp.normalised

    def test_excerpt_is_capped(self):
        log = CHAINED_LOG + "\n".join(f"noise line {i}" for i in range(5000))
        fp = fingerprint(log)
        assert len(fp.excerpt.encode("utf-8")) <= MAX_EXCERPT_BYTES

    def test_truncation_does_not_split_a_character(self):
        log = CHAINED_LOG.replace("kube-proxy", "kübe-pröxy" * 900)
        fingerprint(log).excerpt.encode("utf-8")
