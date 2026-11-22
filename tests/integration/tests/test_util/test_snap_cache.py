#
# Copyright 2026 Canonical, Ltd.
#

import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from test_util import config, harness, tags, util


@pytest.fixture(autouse=True)
def disable_inspection_reports(monkeypatch):
    monkeypatch.setattr(config, "INSPECTION_REPORTS_DIR", None)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_download_k8s_snaps_disabled(monkeypatch, tmp_path):
    monkeypatch.setattr(config, "CACHE_SNAPS", False)
    result = util.download_k8s_snaps(["1.36-classic/candidate"], tmp_path)
    assert result == {}


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.util.run")
def test_download_k8s_snaps_deduplication_and_formats(mock_run, tmp_path):
    channels = [
        "1.36-classic/candidate",
        "/local/snap/path.snap",
        "1.36-classic/candidate",
        "1.36-classic/candidate@5834",
        "5834",
    ]

    result = util.download_k8s_snaps(channels, tmp_path, arch="arm64")

    # Local path should be excluded, duplicates deduplicated
    assert set(result.keys()) == {
        "1.36-classic/candidate",
        "1.36-classic/candidate@5834",
        "5834",
    }
    assert "/local/snap/path.snap" not in result

    # Verify each channel maps to (snap_path, assert_path)
    for ch, (snap_path, assert_path) in result.items():
        assert isinstance(snap_path, Path)
        assert isinstance(assert_path, Path)
        assert snap_path.suffix == ".snap"
        assert assert_path.suffix == ".assert"
        assert snap_path.parent == tmp_path
        assert assert_path.parent == tmp_path

    # Verify run calls and env
    assert mock_run.call_count == 3
    executed_commands = []
    for call_args in mock_run.call_args_list:
        cmd = call_args.args[0] if call_args.args else call_args.kwargs.get("command")
        kwargs = call_args.kwargs
        executed_commands.append(cmd)
        assert kwargs.get("sensitive_kwargs") is True
        env = kwargs.get("env")
        assert env is not None
        assert env.get("UBUNTU_STORE_ARCH") == "arm64"
        assert kwargs.get("timeout") == 300
    # Verify command flags for plain channel, channel@rev, and bare revision
    flat_cmds = [" ".join(c) for c in executed_commands]
    assert any("--channel 1.36-classic/candidate" in c for c in flat_cmds)
    assert any(
        "--channel 1.36-classic/candidate" in c and "--revision 5834" in c
        for c in flat_cmds
    )
    assert any("--revision 5834" in c and "--channel" not in c for c in flat_cmds)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_cached_success(tmp_path):
    instance = MagicMock()
    instance.id = "k8s-0"
    instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"0-3\n", stderr=b""
    )

    snap_file = tmp_path / "k8s_test.snap"
    assert_file = tmp_path / "k8s_test.assert"
    cached = {"1.36-classic/candidate": (snap_file, assert_file)}

    util.setup_k8s_snap(
        instance, tmp_path, "1.36-classic/candidate", cached_snaps=cached
    )

    # Verify push of files to /tmp/<hex>_<name>
    send_calls = instance.send_file.call_args_list
    assert len(send_calls) == 2
    assert send_calls[0].args[0] == str(assert_file)
    remote_assert = send_calls[0].args[1]
    assert remote_assert.startswith("/tmp/")
    assert remote_assert.endswith("_k8s_test.assert")

    assert send_calls[1].args[0] == str(snap_file)
    remote_snap = send_calls[1].args[1]
    assert remote_snap.startswith("/tmp/")
    assert remote_snap.endswith("_k8s_test.snap")

    # Verify snap ack and snap install --classic
    exec_cmds = [c.args[0] for c in instance.exec.call_args_list]
    assert ["snap", "ack", remote_assert] in exec_cmds
    assert ["snap", "install", "--classic", remote_snap] in exec_cmds

    # Verify cleanup in finally:
    cleanup_calls = [
        c
        for c in instance.exec.call_args_list
        if len(c.args[0]) >= 3 and c.args[0][0] == "rm" and c.args[0][1] == "-f"
    ]
    assert len(cleanup_calls) >= 1
    cleanup_cmd = cleanup_calls[0].args[0]
    assert remote_snap in cleanup_cmd
    assert remote_assert in cleanup_cmd
    assert cleanup_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_cached_failure_cleanup_and_retry(tmp_path):
    instance = MagicMock()
    instance.id = "k8s-0"
    snap_file = tmp_path / "k8s_test.snap"
    assert_file = tmp_path / "k8s_test.assert"
    cached = {"1.36-classic/candidate": (snap_file, assert_file)}

    # Fail install once, poll snap changes (showing Doing then Done), then succeed
    def mock_exec(cmd, **kwargs):
        if cmd[:3] == ["snap", "install", "--classic"]:
            if getattr(mock_exec, "installed", False):
                return subprocess.CompletedProcess(
                    args=cmd, returncode=0, stdout="", stderr=""
                )
            mock_exec.installed = True
            raise subprocess.CalledProcessError(
                returncode=10, cmd=cmd, output="change in progress"
            )
        if cmd[:2] == ["snap", "changes"]:
            if getattr(mock_exec, "polled", False):
                return subprocess.CompletedProcess(
                    args=cmd,
                    returncode=0,
                    stdout="1 Done today today Install k8s snap\n",
                    stderr="",
                )
            mock_exec.polled = True
            return subprocess.CompletedProcess(
                args=cmd,
                returncode=0,
                stdout="1 Doing today - Install k8s snap\n",
                stderr="",
            )
        return subprocess.CompletedProcess(
            args=cmd, returncode=0, stdout=b"0-3\n", stderr=b""
        )

    instance.exec.side_effect = mock_exec

    util.setup_k8s_snap(
        instance, tmp_path, "1.36-classic/candidate", cached_snaps=cached, delay_s=0
    )

    # Verify snap changes was checked before retry
    exec_cmds = [c.args[0] for c in instance.exec.call_args_list]
    assert any(c[:2] == ["snap", "changes"] for c in exec_cmds)

    # Verify cleanup occurred
    cleanup_calls = [
        c
        for c in instance.exec.call_args_list
        if len(c.args[0]) >= 3 and c.args[0][0] == "rm" and c.args[0][1] == "-f"
    ]
    assert len(cleanup_calls) >= 1
    assert cleanup_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_fallback_store(tmp_path):
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"0-3\n", stderr=b""
    )

    util.setup_k8s_snap(instance, tmp_path, "1.36-classic/candidate", cached_snaps=None)

    exec_cmds = [c.args[0] for c in instance.exec.call_args_list]
    assert any("snap" in c and "install" in c and "--channel" in c for c in exec_cmds)
    assert not any("rm" in c and "-f" in c for c in exec_cmds)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_cached_success(tmp_path):
    instance = MagicMock()
    instance.id = "k8s-0"
    instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"0-3\n", stderr=b""
    )

    snap_file = tmp_path / "k8s_test.snap"
    assert_file = tmp_path / "k8s_test.assert"
    cached = {"1.36-classic/candidate": (snap_file, assert_file)}

    util.snap_refresh(
        instance,
        "1.36-classic/candidate",
        "--amend",
        "--ignore-validation",
        cached_snaps=cached,
    )

    # Verify push of files to /tmp/
    send_calls = instance.send_file.call_args_list
    assert len(send_calls) == 2
    remote_assert = send_calls[0].args[1]
    assert remote_assert.startswith("/tmp/")
    remote_snap = send_calls[1].args[1]
    assert remote_snap.startswith("/tmp/")

    # MUST use snap install --classic --ignore-validation (strip --amend, NOT snap refresh)
    exec_calls = instance.exec.call_args_list
    install_calls = [
        c
        for c in exec_calls
        if len(c.args[0]) >= 3 and c.args[0][:3] == ["snap", "install", "--classic"]
    ]
    assert len(install_calls) >= 1
    install_cmd = install_calls[0].args[0]
    assert "--amend" not in install_cmd
    assert "--ignore-validation" in install_cmd
    assert remote_snap in install_cmd
    assert install_calls[0].kwargs.get("timeout") == 300

    # Verify cleanup in finally:
    cleanup_calls = [
        c
        for c in instance.exec.call_args_list
        if len(c.args[0]) >= 3 and c.args[0][0] == "rm" and c.args[0][1] == "-f"
    ]
    assert len(cleanup_calls) >= 1
    assert cleanup_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_cached_retry_waiting_changes(tmp_path):
    instance = MagicMock()
    instance.id = "k8s-0"
    snap_file = tmp_path / "k8s_test.snap"
    assert_file = tmp_path / "k8s_test.assert"
    cached = {"1.36-classic/candidate": (snap_file, assert_file)}

    def mock_exec(cmd, **kwargs):
        if cmd[:3] == ["snap", "install", "--classic"]:
            if getattr(mock_exec, "installed", False):
                return subprocess.CompletedProcess(
                    args=cmd, returncode=0, stdout="", stderr=""
                )
            mock_exec.installed = True
            raise subprocess.CalledProcessError(
                returncode=10, cmd=cmd, output="change in progress"
            )
        if cmd[:2] == ["snap", "changes"]:
            return subprocess.CompletedProcess(
                args=cmd,
                returncode=0,
                stdout="1 Done today today Install k8s snap\n",
                stderr="",
            )
        return subprocess.CompletedProcess(
            args=cmd, returncode=0, stdout=b"0-3\n", stderr=b""
        )

    instance.exec.side_effect = mock_exec

    util.snap_refresh(
        instance, "1.36-classic/candidate", cached_snaps=cached, delay_s=0
    )

    exec_cmds = [c.args[0] for c in instance.exec.call_args_list]
    assert any(c[:2] == ["snap", "changes"] for c in exec_cmds)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_fallback_store():
    instance = MagicMock()
    instance.id = "k8s-0"
    instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"0-3\n", stderr=b""
    )

    util.snap_refresh(instance, "1.36-classic/candidate", "--amend", cached_snaps=None)

    exec_cmds = [c.args[0] for c in instance.exec.call_args_list]
    assert any("snap" in c and "refresh" in c and "--amend" in c for c in exec_cmds)


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("urllib.request.urlopen")
def test_fetch_text_timeout_and_retry(mock_urlopen):
    # First attempt times out with TimeoutError, second succeeds
    mock_resp = MagicMock()
    mock_resp.__enter__.return_value.read.return_value = b"success"
    mock_urlopen.side_effect = [TimeoutError("connection timed out"), mock_resp]

    result = util.fetch_text(
        "https://example.com/test", min_wait_s=0.01, max_wait_s=0.02
    )
    assert result == "success"
    assert mock_urlopen.call_count == 2
    # Verify timeout=30.0 was passed
    assert mock_urlopen.call_args_list[0].kwargs.get("timeout") == 30.0


@pytest.mark.tags(tags.PULL_REQUEST)
def test_instance_arch_timeout_and_retry():
    mock_harness = MagicMock()
    instance = harness.base.Instance(mock_harness, "k8s-0")

    # First attempt times out, second succeeds
    mock_harness.exec.side_effect = [
        subprocess.TimeoutExpired(cmd=["dpkg", "--print-architecture"], timeout=60),
        subprocess.CompletedProcess(
            args=["dpkg", "--print-architecture"], returncode=0, stdout="amd64\n"
        ),
    ]

    arch = instance.arch
    assert arch == "amd64"
    assert mock_harness.exec.call_count == 2
    assert mock_harness.exec.call_args_list[0].kwargs.get("timeout") == 60


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.util.run")
def test_download_k8s_snaps_timeout_retry_and_cleanup(mock_run, tmp_path):
    snap_partial = tmp_path / "k8s_1.36-classic_candidate.snap"
    assert_partial = tmp_path / "k8s_1.36-classic_candidate.assert"

    def side_effect(cmd, **kwargs):
        if mock_run.call_count == 1:
            snap_partial.write_text("partial-snap")
            assert_partial.write_text("partial-assert")
            raise subprocess.TimeoutExpired(cmd=cmd, timeout=300)
        snap_partial.write_text("complete-snap")
        assert_partial.write_text("complete-assert")
        return subprocess.CompletedProcess(
            args=cmd, returncode=0, stdout=b"", stderr=b""
        )

    mock_run.side_effect = side_effect

    result = util.download_k8s_snaps(
        ["1.36-classic/candidate"], tmp_path, retries=2, delay_s=0, timeout=300
    )

    assert "1.36-classic/candidate" in result
    assert mock_run.call_count == 2
    assert mock_run.call_args_list[0].kwargs.get("timeout") == 300
    assert snap_partial.read_text() == "complete-snap"
    assert assert_partial.read_text() == "complete-assert"


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.util.run")
def test_download_k8s_snaps_exhausted_retries_cleanup_and_raises(mock_run, tmp_path):
    snap_partial = tmp_path / "k8s_1.36-classic_candidate.snap"

    def side_effect(cmd, **kwargs):
        snap_partial.write_text("partial-snap")
        raise subprocess.TimeoutExpired(cmd=cmd, timeout=300)

    mock_run.side_effect = side_effect

    with pytest.raises(subprocess.TimeoutExpired):
        util.download_k8s_snaps(
            ["1.36-classic/candidate"], tmp_path, retries=2, delay_s=0, timeout=300
        )

    assert mock_run.call_count == 2
    assert not snap_partial.exists()
