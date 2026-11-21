#
# Copyright 2026 Canonical, Ltd.
#
import logging
import subprocess
from unittest.mock import MagicMock

import pytest
from tenacity import RetryError
from test_util import harness, tags, util  # noqa: F401


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_services_stopped_polls_until_inactive():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["snap", "services", "k8s"],
            returncode=0,
            stdout=(
                "Service        Startup  Current   Notes\n"
                "k8s.k8sd       enabled  active    -\n"
                "k8s.kubelet    enabled  active    -\n"
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "services", "k8s"],
            returncode=0,
            stdout=(
                "Service        Startup  Current   Notes\n"
                "k8s.k8sd       enabled  active    -\n"
                "k8s.kubelet    enabled  inactive  -\n"
            ),
            stderr="",
        ),
    ]

    util.wait_for_services_stopped(
        instance, exclude_services=["k8sd"], retries=2, delay_s=0
    )

    assert instance.exec.call_count == 2


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_services_stopped_raises_when_still_active():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["snap", "services", "k8s"],
        returncode=0,
        stdout=(
            "Service        Startup  Current   Notes\n"
            "k8s.kubelet    enabled  active    -\n"
        ),
        stderr="",
    )

    with pytest.raises(RetryError):
        util.wait_for_services_stopped(instance, retries=2, delay_s=0)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_ports_available_polls_until_unbound():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["ss", "-tuln"],
            returncode=0,
            stdout=(
                "Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n"
                "tcp LISTEN 0 4096 0.0.0.0:6443 0.0.0.0:*\n"
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["ss", "-tuln"],
            returncode=0,
            stdout="Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n",
            stderr="",
        ),
    ]

    util.wait_for_ports_available(instance, ports=[6443], retries=2, delay_s=0)

    assert instance.exec.call_count == 2


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_ports_available_raises_when_still_bound():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["ss", "-tuln"],
        returncode=0,
        stdout=(
            "Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n"
            "tcp LISTEN 0 4096 [::]:6443 [::]:*\n"
        ),
        stderr="",
    )

    with pytest.raises(RetryError):
        util.wait_for_ports_available(instance, ports=[6443], retries=2, delay_s=0)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_stubbornly_until_omits_command_streams_from_retry_logs(caplog):
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["k8s", "status"],
            returncode=0,
            stdout="sentinel-poll-stdout\n",
            stderr="sentinel-poll-stderr\n",
        ),
        subprocess.CompletedProcess(
            args=["k8s", "status"],
            returncode=0,
            stdout="sentinel-poll-stdout\n",
            stderr="sentinel-poll-stderr\n",
        ),
        subprocess.CompletedProcess(
            args=["k8s", "status"],
            returncode=0,
            stdout="cluster ready\n",
            stderr="",
        ),
    ]

    with caplog.at_level(logging.DEBUG):
        util.stubbornly(retries=3, delay_s=0).on(instance).until(
            lambda p: "ready" in p.stdout
        ).exec(["k8s", "status"], text=True)

    assert "Attempt 1/3 failed" in caplog.text
    assert "Attempt 2/3 failed" in caplog.text
    assert "rc=0" in caplog.text
    assert "sentinel-poll-stdout" not in caplog.text
    assert "sentinel-poll-stderr" not in caplog.text


@pytest.mark.tags(tags.PULL_REQUEST)
def test_stubbornly_until_includes_truncated_streams_on_exhaustion():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["k8s", "kubectl", "get", "pods", "-o", "json"],
        returncode=0,
        stdout=b"HEAD_MARKER_" + (b"x" * 2000) + b"_TAIL_MARKER",
        stderr=b"short-err",
    )

    with pytest.raises(RetryError) as exc_info:
        util.stubbornly(retries=2, delay_s=0).on(instance).until(lambda p: False).exec(
            ["k8s", "kubectl", "get", "pods", "-o", "json"]
        )

    cause = exc_info.value.last_attempt.exception()
    assert isinstance(cause, AssertionError)
    msg = str(cause)
    assert "rc=0" in msg
    assert "HEAD_MARKER_" in msg
    assert "_TAIL_MARKER" in msg
    assert "...[truncated 1024 chars]..." in msg
    assert "short-err" in msg
    assert len(msg) < 1200


@pytest.mark.tags(tags.PULL_REQUEST)
def test_stubbornly_called_process_error_truncates_logged_streams(caplog):
    instance = MagicMock()
    raw_stdout = b"OUT_START_" + (b"a" * 2000) + b"_OUT_END"
    raw_stderr = b"ERR_START_" + (b"b" * 2000) + b"_ERR_END"
    instance.exec.side_effect = subprocess.CalledProcessError(
        returncode=1,
        cmd=["k8s", "status"],
        output=raw_stdout,
        stderr=raw_stderr,
    )

    with caplog.at_level(logging.WARNING), pytest.raises(RetryError) as exc_info:
        util.stubbornly(retries=1, delay_s=0).on(instance).exec(["k8s", "status"])

    assert "OUT_START_" in caplog.text
    assert "_OUT_END" in caplog.text
    assert "...[truncated 1018 chars]..." in caplog.text
    assert "ERR_START_" in caplog.text
    assert "_ERR_END" in caplog.text
    assert ("a" * 600) not in caplog.text
    assert ("b" * 600) not in caplog.text
    cause = exc_info.value.last_attempt.exception()
    assert cause.stdout == raw_stdout
    assert cause.stderr == raw_stderr


@pytest.mark.tags(tags.PULL_REQUEST)
def test_stubbornly_timeout_expired_truncates_logged_streams(caplog):
    instance = MagicMock()
    instance.exec.side_effect = subprocess.TimeoutExpired(
        cmd=["k8s", "status"],
        timeout=300,
        output=b"TO_START_" + (b"t" * 2000) + b"_TO_END",
        stderr=b"TE_START_" + (b"e" * 2000) + b"_TE_END",
    )

    with caplog.at_level(logging.WARNING), pytest.raises(RetryError):
        util.stubbornly(retries=1, delay_s=0).on(instance).exec(
            ["k8s", "status"], timeout=300
        )

    assert "TO_START_" in caplog.text
    assert "_TO_END" in caplog.text
    assert "...[truncated 1016 chars]..." in caplog.text
    assert "TE_START_" in caplog.text
    assert "_TE_END" in caplog.text
    assert ("t" * 600) not in caplog.text


@pytest.mark.tags(tags.PULL_REQUEST)
def test_truncate_output_preserves_exact_boundary():
    exact = "a" * 1000
    assert util._truncate_output(exact) == exact

    over_by_one = ("a" * 500) + "M" + ("b" * 500)
    expected = ("a" * 500) + "...[truncated 1 chars]..." + ("b" * 500)
    assert util._truncate_output(over_by_one) == expected
