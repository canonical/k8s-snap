#
# Copyright 2026 Canonical, Ltd.
#

import subprocess
from unittest.mock import MagicMock

import pytest
from tenacity import RetryError
from test_util import tags, util


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_changes_polls_until_settled():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Doing    yesterday at 14:44 CET  -                       Install snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Undoing  yesterday at 14:44 CET  -                       Install snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Do       yesterday at 14:44 CET  -                       Install snap "k8s"\n'
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=(
                "ID   Status   Spawn                   Ready                   Summary\n"
                '1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Install snap "k8s"\n'
                '2    Error    yesterday at 14:40 CET  yesterday at 14:41 CET  Refresh snap "core22"\n'
            ),
            stderr="",
        ),
    ]

    util.wait_for_snap_changes(instance, retries=5, delay_s=0)

    assert instance.exec.call_count == 4


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_changes_ignores_status_words_in_summary():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["snap", "changes"],
        returncode=0,
        stdout=(
            "ID   Status   Spawn                   Ready                   Summary\n"
            "1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Doing some maintenance and Do work\n"
        ),
        stderr="",
    )

    util.wait_for_snap_changes(instance, retries=2, delay_s=0)

    assert instance.exec.call_count == 1


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_changes_handles_bytes_stdout_and_empty_output():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=(
                b"ID   Status   Spawn                   Ready                   Summary\n"
                b'1    Done     yesterday at 14:44 CET  yesterday at 14:45 CET  Install snap "k8s"\n'
            ),
            stderr=b"",
        ),
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=b"No changes found.\n",
            stderr=b"",
        ),
    ]

    util.wait_for_snap_changes(instance, retries=2, delay_s=0)
    util.wait_for_snap_changes(instance, retries=2, delay_s=0)

    assert instance.exec.call_count == 2


@pytest.mark.tags(tags.PULL_REQUEST)
def test_wait_for_snap_changes_raises_on_timeout():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["snap", "changes"],
        returncode=0,
        stdout=(
            "ID   Status   Spawn                   Ready                   Summary\n"
            '1    Doing    yesterday at 14:44 CET  -                       Install snap "k8s"\n'
        ),
        stderr="",
    )

    with pytest.raises(RetryError):
        util.wait_for_snap_changes(instance, retries=2, delay_s=0)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_waits_for_changes_across_timeout_retry():
    instance = MagicMock()
    instance.arch = "amd64"
    instance.exec.side_effect = [
        # configure_cpuset_controller
        subprocess.CompletedProcess(
            args=["cat", "/sys/devices/system/cpu/online"],
            returncode=0,
            stdout="0-3\n",
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["systemctl"], returncode=0, stdout="", stderr=""
        ),
        # Attempt 1: wait_for_snap_changes
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout="ID   Status  Summary\n1    Done    Setup\n",
            stderr="",
        ),
        # Attempt 1: snap install times out
        subprocess.TimeoutExpired(
            cmd=[
                "snap",
                "install",
                "--classic",
                "k8s",
                "--channel",
                "latest/edge/classic",
            ],
            timeout=600,
        ),
        # Attempt 2: wait_for_snap_changes - background install is still Doing
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=("ID   Status  Summary\n" '2    Doing   Install snap "k8s"\n'),
            stderr="",
        ),
        # Attempt 2: wait_for_snap_changes - background install is now Done
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=("ID   Status  Summary\n" '2    Done    Install snap "k8s"\n'),
            stderr="",
        ),
        # Attempt 2: snap install runs and succeeds (already installed)
        subprocess.CompletedProcess(
            args=[
                "snap",
                "install",
                "--classic",
                "k8s",
                "--channel",
                "latest/edge/classic",
            ],
            returncode=0,
            stdout="snap \"k8s\" is already installed, see 'snap help refresh'\n",
            stderr="",
        ),
        # connect_interfaces: init.sh
        subprocess.CompletedProcess(
            args=["/snap/k8s/current/k8s/hack/init.sh"],
            returncode=0,
            stdout="",
            stderr="",
        ),
    ]

    util.setup_k8s_snap(
        instance,
        snap="latest/edge/classic",
        connect_interfaces=True,
        retries=2,
        delay_s=0,
        timeout=600,
    )

    # Verify snap changes was checked before each attempt
    exec_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert exec_cmds == [
        ["cat", "/sys/devices/system/cpu/online"],
        ["systemctl", "set-property", "--runtime", "system.slice", "AllowedCPUs=0-3"],
        ["snap", "changes"],
        ["snap", "install", "--classic", "k8s", "--channel", "latest/edge/classic"],
        ["snap", "changes"],
        ["snap", "changes"],
        ["snap", "install", "--classic", "k8s", "--channel", "latest/edge/classic"],
        ["/snap/k8s/current/k8s/hack/init.sh"],
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_waits_for_changes_across_error_retry():
    instance = MagicMock()
    instance.exec.side_effect = [
        # Attempt 1: wait_for_snap_changes
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout="ID   Status  Summary\n1    Done    Setup\n",
            stderr="",
        ),
        # Attempt 1: snap refresh fails with change in progress
        subprocess.CalledProcessError(
            returncode=10,
            cmd=[
                "snap",
                "refresh",
                "k8s",
                "--channel",
                "1.32-classic/stable",
                "--classic",
            ],
            stderr='error: snap "k8s" has "install-snap" change in progress\n',
        ),
        # Attempt 2: wait_for_snap_changes - background change is still Doing
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=("ID   Status  Summary\n" '2    Doing   Install snap "k8s"\n'),
            stderr="",
        ),
        # Attempt 2: wait_for_snap_changes - background change is Done
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout=("ID   Status  Summary\n" '2    Done    Install snap "k8s"\n'),
            stderr="",
        ),
        # Attempt 2: snap refresh succeeds
        subprocess.CompletedProcess(
            args=[
                "snap",
                "refresh",
                "k8s",
                "--channel",
                "1.32-classic/stable",
                "--classic",
            ],
            returncode=0,
            stdout='snap "k8s" has no updates available\n',
            stderr="",
        ),
    ]

    util.snap_refresh(
        instance,
        channel="1.32-classic/stable",
        retries=2,
        delay_s=0,
        timeout=600,
    )

    exec_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert exec_cmds == [
        ["snap", "changes"],
        ["snap", "refresh", "k8s", "--channel", "1.32-classic/stable", "--classic"],
        ["snap", "changes"],
        ["snap", "changes"],
        ["snap", "refresh", "k8s", "--channel", "1.32-classic/stable", "--classic"],
    ]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_handles_already_installed_error_gracefully():
    instance = MagicMock()
    instance.arch = "amd64"
    instance.exec.side_effect = [
        # configure_cpuset_controller
        subprocess.CompletedProcess(
            args=["cat", "/sys/devices/system/cpu/online"],
            returncode=0,
            stdout="0-3\n",
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["systemctl"], returncode=0, stdout="", stderr=""
        ),
        # wait_for_snap_changes
        subprocess.CompletedProcess(
            args=["snap", "changes"],
            returncode=0,
            stdout="ID   Status  Summary\n1    Done    Setup\n",
            stderr="",
        ),
        # snap install raises CalledProcessError with "already installed"
        subprocess.CalledProcessError(
            returncode=1,
            cmd=[
                "snap",
                "install",
                "--classic",
                "k8s",
                "--channel",
                "latest/edge/classic",
            ],
            stderr="snap \"k8s\" is already installed, see 'snap help refresh'\n",
        ),
        # connect_interfaces: init.sh
        subprocess.CompletedProcess(
            args=["/snap/k8s/current/k8s/hack/init.sh"],
            returncode=0,
            stdout="",
            stderr="",
        ),
    ]

    util.setup_k8s_snap(
        instance,
        snap="latest/edge/classic",
        connect_interfaces=True,
        retries=2,
        delay_s=0,
        timeout=600,
    )

    exec_cmds = [call_args[0][0] for call_args in instance.exec.call_args_list]
    assert exec_cmds == [
        ["cat", "/sys/devices/system/cpu/online"],
        ["systemctl", "set-property", "--runtime", "system.slice", "AllowedCPUs=0-3"],
        ["snap", "changes"],
        ["snap", "install", "--classic", "k8s", "--channel", "latest/edge/classic"],
        ["/snap/k8s/current/k8s/hack/init.sh"],
    ]
