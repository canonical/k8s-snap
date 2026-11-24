#
# Copyright 2026 Canonical, Ltd.
#
import subprocess
from unittest.mock import MagicMock

from test_util import util


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


def test_snap_revert_uses_bounded_timeout_and_retries_after_timeout():
    instance = MagicMock()
    cmd = ["snap", "revert", "k8s"]
    instance.exec.side_effect = [
        subprocess.TimeoutExpired(cmd=cmd, timeout=300),
        subprocess.CompletedProcess(
            args=["snap", "changes", "k8s"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Revert snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=cmd,
            returncode=0,
            stdout="k8s reverted",
            stderr="",
        ),
    ]

    util.snap_revert(instance, "k8s", retries=2, delay_s=0)

    assert instance.exec.call_count == 3
    executed_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert executed_cmds == [
        cmd,
        ["snap", "changes", "k8s"],
        cmd,
    ]
    assert instance.exec.call_args_list[0][1].get("timeout") == 300
    assert instance.exec.call_args_list[2][1].get("timeout") == 300
