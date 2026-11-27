#
# Copyright 2026 Canonical, Ltd.
#
import base64
import json
import logging
import subprocess
from unittest.mock import MagicMock

import pytest
from tenacity import RetryError
from test_util import util


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


def test_truncate_output_preserves_exact_boundary():
    exact = "a" * 1000
    assert util._truncate_output(exact) == exact

    over_by_one = ("a" * 500) + "M" + ("b" * 500)
    expected = ("a" * 500) + "...[truncated 1 chars]..." + ("b" * 500)
    assert util._truncate_output(over_by_one) == expected


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


def test_is_daemonset_ready_polls_until_scheduled_and_ready():
    control_node = MagicMock()
    ds_ready_lagging = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 2,
                        "numberReady": 1,
                        "updatedNumberScheduled": 2,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                }
            ]
        }
    )
    ds_updated_lagging = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 2,
                        "numberReady": 2,
                        "updatedNumberScheduled": 1,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                }
            ]
        }
    )
    ds_ready = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 2,
                        "numberReady": 2,
                        "updatedNumberScheduled": 2,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                }
            ]
        }
    )
    pod_unready = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium-cp1"},
                    "status": {"conditions": [{"type": "Ready", "status": "False"}]},
                }
            ]
        }
    )
    pod_ready = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium-cp1"},
                    "status": {"conditions": [{"type": "Ready", "status": "True"}]},
                }
            ]
        }
    )

    control_node.exec.side_effect = [
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_ready_lagging),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_updated_lagging),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_ready),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=pod_unready),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_ready),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=pod_ready),
    ]

    assert not util.is_daemonset_ready(control_node, node_name="cp1")
    assert not util.is_daemonset_ready(control_node, node_name="cp1")
    assert not util.is_daemonset_ready(control_node, node_name="cp1")
    assert util.is_daemonset_ready(control_node, node_name="cp1")


def test_is_daemonset_ready_skips_zero_desired_daemonset():
    control_node = MagicMock()
    ds_data = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "gpu-operator"},
                    "status": {
                        "desiredNumberScheduled": 0,
                        "numberReady": 0,
                        "updatedNumberScheduled": 0,
                    },
                    "spec": {"selector": {"matchLabels": {"app": "gpu-operator"}}},
                },
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 1,
                        "numberReady": 1,
                        "updatedNumberScheduled": 1,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                },
            ]
        }
    )
    pod_ready = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium-cp1"},
                    "status": {"conditions": [{"type": "Ready", "status": "True"}]},
                }
            ]
        }
    )
    control_node.exec.side_effect = [
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_data),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=pod_ready),
    ]
    assert util.is_daemonset_ready(control_node, node_name="cp1")


def test_is_daemonset_ready_blocks_when_updated_less_than_desired():
    control_node = MagicMock()
    ds_lagging = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 2,
                        "numberReady": 2,
                        "updatedNumberScheduled": 1,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                }
            ]
        }
    )
    control_node.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=ds_lagging
    )
    assert not util.is_daemonset_ready(control_node, node_name="cp1")


def test_is_daemonset_ready_returns_false_when_pod_missing():
    control_node = MagicMock()
    ds_ready = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "cilium"},
                    "status": {
                        "desiredNumberScheduled": 2,
                        "numberReady": 2,
                        "updatedNumberScheduled": 2,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "cilium"}}},
                }
            ]
        }
    )
    pods_empty = json.dumps({"items": []})

    control_node.exec.side_effect = [
        subprocess.CompletedProcess(args=[], returncode=0, stdout=ds_ready),
        subprocess.CompletedProcess(args=[], returncode=0, stdout=pods_empty),
    ]

    assert not util.is_daemonset_ready(control_node, node_name="cp1")


def test_is_daemonset_ready_returns_false_when_no_daemonsets_found():
    control_node = MagicMock()
    control_node.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout='{"items": []}'
    )
    assert not util.is_daemonset_ready(control_node)


def test_is_daemonset_ready_returns_false_when_cilium_missing_in_kube_system():
    control_node = MagicMock()
    ds_other = json.dumps(
        {
            "items": [
                {
                    "metadata": {"name": "other"},
                    "status": {
                        "desiredNumberScheduled": 1,
                        "numberReady": 1,
                        "updatedNumberScheduled": 1,
                    },
                    "spec": {"selector": {"matchLabels": {"k8s-app": "other"}}},
                }
            ]
        }
    )
    control_node.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=ds_other
    )
    assert not util.is_daemonset_ready(control_node)


def _join_token(**fields) -> str:
    return base64.b64encode(json.dumps(fields).encode()).decode()


def test_join_cluster_waits_for_control_plane_node(monkeypatch):
    instance = MagicMock()
    instance.id = "cp1"
    wait_mock = MagicMock()
    monkeypatch.setattr(util, "wait_until_k8s_ready", wait_mock)
    monkeypatch.setattr(util, "hostname", lambda _: "cp1")

    util.join_cluster(instance, _join_token(secret="s", fingerprint="f"))

    wait_mock.assert_called_once_with(
        instance,
        [instance],
        node_names={"cp1": "cp1"},
        wait_for_daemonsets=True,
    )


def test_join_cluster_skips_wait_for_worker_or_when_disabled(monkeypatch):
    instance = MagicMock()
    wait_mock = MagicMock()
    monkeypatch.setattr(util, "wait_until_k8s_ready", wait_mock)

    util.join_cluster(instance, _join_token(secret="s", fingerprint="f", _="m!!"))
    util.join_cluster(instance, _join_token(secret="s"), wait_ready=False)

    wait_mock.assert_not_called()
