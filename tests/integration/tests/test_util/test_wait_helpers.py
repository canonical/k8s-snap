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
