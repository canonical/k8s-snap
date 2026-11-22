#
# Copyright 2026 Canonical, Ltd.
#
import logging
import subprocess
from unittest.mock import MagicMock

import pytest
from tenacity import RetryError
from test_util import tags, util


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


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_change_happy_path_no_snap_changes():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["snap", "install", "--classic", "k8s"],
        returncode=0,
        stdout="k8s installed",
        stderr="",
    )

    cmd = ["snap", "install", "--classic", "k8s"]
    res = (
        util.stubbornly(
            retries=3,
            delay_s=0,
            after=util.wait_for_snap_change_hook(
                instance, max_attempts=3, retries=2, delay_s=0
            ),
        )
        .on(instance)
        .exec(cmd)
    )

    assert res.returncode == 0
    assert instance.exec.call_count == 1
    assert instance.exec.call_args[0][0] == cmd


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_change_waits_for_inflight_change_before_retry():
    instance = MagicMock()
    cmd = ["snap", "install", "--classic", "k8s"]
    instance.exec.side_effect = [
        subprocess.TimeoutExpired(cmd=cmd, timeout=300),
        subprocess.CompletedProcess(
            args=["snap", "changes", "k8s"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Doing    yesterday at 14:44 CET  -                       Install snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "changes", "k8s"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Install snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=cmd,
            returncode=0,
            stdout="k8s installed",
            stderr="",
        ),
    ]

    res = (
        util.stubbornly(
            retries=2,
            delay_s=0,
            after=util.wait_for_snap_change_hook(
                instance, max_attempts=2, retries=2, delay_s=0
            ),
        )
        .on(instance)
        .exec(cmd)
    )

    assert res.returncode == 0
    assert instance.exec.call_count == 4
    executed_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert executed_cmds == [
        cmd,
        ["snap", "changes", "k8s"],
        ["snap", "changes", "k8s"],
        cmd,
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_uses_bounded_timeout_and_retries_after_timeout():
    instance = MagicMock()
    cmd = [
        "snap",
        "refresh",
        "k8s",
        "--channel",
        "1.36-classic/candidate",
        "--classic",
        "--amend",
    ]
    instance.exec.side_effect = [
        subprocess.TimeoutExpired(cmd=cmd, timeout=300),
        subprocess.CompletedProcess(
            args=["snap", "changes", "k8s"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Refresh snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=cmd,
            returncode=0,
            stdout="k8s refreshed",
            stderr="",
        ),
    ]

    util.snap_refresh(
        instance,
        "1.36-classic/candidate",
        "--amend",
        retries=2,
        delay_s=0,
    )

    assert instance.exec.call_count == 3
    executed_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert executed_cmds == [
        cmd,
        ["snap", "changes", "k8s"],
        cmd,
    ]
    assert instance.exec.call_args_list[0][1].get("timeout") == 300
    assert instance.exec.call_args_list[2][1].get("timeout") == 300
