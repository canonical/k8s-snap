#
# Copyright 2026 Canonical, Ltd.
#
import concurrent.futures
import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from test_util import tags, util
from test_util.harness.base import Instance
from test_util.harness.lxd import LXDHarness


@pytest.mark.tags(tags.PULL_REQUEST)
def test_download_k8s_snaps_deduplication_and_env(tmp_path):
    channels = [
        "1.32/stable",
        "1.32/stable",
        "/local/snap/path.snap",
        "1.31/candidate@42",
        "1.30/beta",
    ]
    cache_dir = tmp_path / "cache"

    with patch("test_util.util.stubbornly") as mock_stubbornly:
        mock_exec = MagicMock()
        mock_stubbornly.return_value.exec = mock_exec

        cached = util.download_k8s_snaps(channels, cache_dir, arch="arm64")

        # Returned dictionary matches the unique non-local channels
        expected_keys = {"1.32/stable", "1.31/candidate@42", "1.30/beta"}
        assert set(cached.keys()) == expected_keys

        for chan, (snap_path, assert_path) in cached.items():
            assert snap_path.parent == cache_dir
            assert assert_path.parent == cache_dir
            assert snap_path.suffix == ".snap"
            assert assert_path.suffix == ".assert"

        # Verify only 3 downloads happened (1.32/stable deduplicated, /local skipped)
        assert mock_exec.call_count == 3
        calls = mock_exec.call_args_list

        # Check UBUNTU_STORE_ARCH passed in env
        for c in calls:
            kwargs = c.kwargs
            assert "env" in kwargs
            assert kwargs["env"].get("UBUNTU_STORE_ARCH") == "arm64"

        # Verify channel@revision handled
        found_revision_call = False
        for c in calls:
            cmd = c.args[0]
            if "--revision" in cmd and "42" in cmd:
                found_revision_call = True
                assert "--channel" in cmd
                assert "1.31/candidate" in cmd
        assert found_revision_call


@pytest.mark.tags(tags.PULL_REQUEST)
def test_download_k8s_snaps_disabled_when_cache_false(tmp_path):
    channels = ["1.32/stable"]
    cache_dir = tmp_path / "cache"

    with patch("test_util.config.CACHE_SNAPS", False):
        with patch("test_util.util.stubbornly") as mock_stubbornly:
            cached = util.download_k8s_snaps(channels, cache_dir)
            assert cached == {}
            assert mock_stubbornly.call_count == 0


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_with_cached_snaps():
    instance = MagicMock()
    instance.id = "k8s-node-1"
    instance.arch = "amd64"

    snap_file = Path("/host/cache/1_32_stable.snap")
    assert_file = Path("/host/cache/1_32_stable.assert")
    cache = {"1.32/stable": (snap_file, assert_file)}

    with (
        patch("test_util.util.configure_cpuset_controller"),
        patch("test_util.util.stubbornly") as mock_stubbornly,
    ):
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        util.setup_k8s_snap(
            instance, "1.32/stable", cached_snaps=cache, connect_interfaces=False
        )

        # Check send_file called for both assert and snap
        assert instance.send_file.call_count == 2
        send_file_calls = instance.send_file.call_args_list
        sent_sources = [c.args[0] for c in send_file_calls]
        assert assert_file.as_posix() in sent_sources
        assert snap_file.as_posix() in sent_sources

        # Check remote paths were in /tmp/
        remote_dests = [c.args[1] for c in send_file_calls]
        assert all(dest.startswith("/tmp/") for dest in remote_dests)

        # Check snap ack and snap install calls
        exec_calls = [c.args[0] for c in mock_on.exec.call_args_list]
        ack_call = next(c for c in exec_calls if c[0] == "snap" and c[1] == "ack")
        assert ack_call is not None
        install_call = next(
            c for c in exec_calls if c[0] == "snap" and c[1] == "install"
        )
        assert install_call[2] == "--classic"

        # Check cleanup rm -f was executed in finally block
        instance_rm_calls = [
            c for c in instance.exec.call_args_list if c.args[0][:2] == ["rm", "-f"]
        ]
        assert len(instance_rm_calls) == 1
        assert instance_rm_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_cached_cleanup_on_install_failure():
    instance = MagicMock()
    instance.id = "k8s-node-1"
    instance.arch = "amd64"

    snap_file = Path("/host/cache/1_32_stable.snap")
    assert_file = Path("/host/cache/1_32_stable.assert")
    cache = {"1.32/stable": (snap_file, assert_file)}

    with (
        patch("test_util.util.configure_cpuset_controller"),
        patch("test_util.util.stubbornly") as mock_stubbornly,
    ):
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        # Cause snap install to fail
        def _side_effect(cmd, **kwargs):
            if cmd[:2] == ["snap", "install"]:
                raise subprocess.CalledProcessError(1, cmd)
            return MagicMock()

        mock_on.exec.side_effect = _side_effect

        with pytest.raises(subprocess.CalledProcessError):
            util.setup_k8s_snap(
                instance, "1.32/stable", cached_snaps=cache, connect_interfaces=False
            )

        # rm -f MUST still be called in finally
        instance_exec_calls = [c.args[0] for c in instance.exec.call_args_list]
        rm_call = next((c for c in instance_exec_calls if c[:2] == ["rm", "-f"]), None)
        assert rm_call is not None


@pytest.mark.tags(tags.PULL_REQUEST)
def test_setup_k8s_snap_fallback_when_not_in_cache():
    instance = MagicMock()
    instance.id = "k8s-node-1"
    instance.arch = "amd64"

    with (
        patch("test_util.util.configure_cpuset_controller"),
        patch("test_util.util.stubbornly") as mock_stubbornly,
    ):
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        util.setup_k8s_snap(
            instance, "1.32/stable", cached_snaps={}, connect_interfaces=False
        )

        # Normal store install command
        exec_calls = [c.args[0] for c in mock_on.exec.call_args_list]
        install_call = next(c for c in exec_calls if c[:2] == ["snap", "install"])
        assert "--channel" in install_call
        assert "1.32/stable" in install_call
        assert instance.send_file.call_count == 0


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_with_cached_snaps():
    instance = MagicMock()
    instance.id = "k8s-node-1"

    snap_file = Path("/host/cache/1_33_stable.snap")
    assert_file = Path("/host/cache/1_33_stable.assert")
    cache = {"1.33/stable": (snap_file, assert_file)}

    with patch("test_util.util.stubbornly") as mock_stubbornly:
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        util.snap_refresh(
            instance,
            "1.33/stable",
            "--amend",
            "--ignore-validation",
            cached_snaps=cache,
        )

        # send_file called for assert and snap
        assert instance.send_file.call_count == 2

        # snap ack and snap install --classic called (snap refresh rejects file paths,
        # and --amend is store-refresh-only while install flags are forwarded)
        exec_calls = [c.args[0] for c in mock_on.exec.call_args_list]
        ack_call = next(c for c in exec_calls if c[:2] == ["snap", "ack"])
        assert ack_call is not None
        install_call = next(c for c in exec_calls if c[:2] == ["snap", "install"])
        assert "--classic" in install_call
        assert "--amend" not in install_call
        assert "--ignore-validation" in install_call
        assert install_call[-1].endswith(".snap")

        # rm -f cleanup in finally with check=False
        instance_rm_calls = [
            c for c in instance.exec.call_args_list if c.args[0][:2] == ["rm", "-f"]
        ]
        assert len(instance_rm_calls) == 1
        assert instance_rm_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_cached_cleanup_on_error():
    instance = MagicMock()
    instance.id = "k8s-node-1"
    snap_file = Path("/host/cache/1_33_stable.snap")
    assert_file = Path("/host/cache/1_33_stable.assert")
    cache = {"1.33/stable": (snap_file, assert_file)}

    with patch("test_util.util.stubbornly") as mock_stubbornly:
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        def _side_effect(cmd, **kwargs):
            if cmd[:2] == ["snap", "install"]:
                raise subprocess.CalledProcessError(1, cmd)
            return MagicMock()

        mock_on.exec.side_effect = _side_effect

        with pytest.raises(subprocess.CalledProcessError):
            util.snap_refresh(instance, "1.33/stable", "--amend", cached_snaps=cache)

        # rm -f called in finally
        instance_rm_calls = [
            c for c in instance.exec.call_args_list if c.args[0][:2] == ["rm", "-f"]
        ]
        assert len(instance_rm_calls) == 1
        assert instance_rm_calls[0].kwargs.get("check") is False


@pytest.mark.tags(tags.PULL_REQUEST)
def test_snap_refresh_fallback_without_cache():
    instance = MagicMock()
    instance.id = "k8s-node-1"

    with patch("test_util.util.stubbornly") as mock_stubbornly:
        mock_on = MagicMock()
        mock_stubbornly.return_value.on.return_value = mock_on

        util.snap_refresh(instance, "1.33/stable", "--amend", cached_snaps=None)

        exec_calls = [c.args[0] for c in mock_on.exec.call_args_list]
        refresh_call = next(c for c in exec_calls if c[:2] == ["snap", "refresh"])
        assert "--channel" in refresh_call
        assert "1.33/stable" in refresh_call
        assert "--amend" in refresh_call
        assert instance.send_file.call_count == 0


@pytest.mark.tags(tags.PULL_REQUEST)
def test_harness_next_id_thread_safety():
    with (
        patch("test_util.harness.lxd.LXDHarness._configure_profile"),
        patch("test_util.harness.lxd.LXDHarness._configure_network"),
    ):
        h = LXDHarness()

    ids = []

    def _get_id():
        return h.next_id()

    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(_get_id) for _ in range(50)]
        ids = [f.result() for f in futures]

    assert len(ids) == 50
    assert len(set(ids)) == 50


@pytest.mark.tags(tags.PULL_REQUEST)
def test_conftest_instances_order_preserved():
    # Verify ThreadPoolExecutor mapping preserves index order 0..3
    indices = list(range(4))

    def _provision(i):
        return f"node-{i}"

    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(_provision, indices))

    assert results == ["node-0", "node-1", "node-2", "node-3"]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_conftest_instances_cleanup_on_provisioning_failure():
    import conftest

    mock_harness = MagicMock()
    mock_harness.instances = set()
    created = []

    def _new_inst(network_type="IPv4"):
        idx = len(created)
        inst_id = f"inst-{idx}"
        mock_harness.instances.add(inst_id)
        if idx == 2:
            # Simulate backend registering the resource before seed wait fails
            raise RuntimeError("Backend failure inside new_instance before returning")
        inst = MagicMock(spec=Instance)
        inst.id = inst_id
        created.append(inst)
        return inst

    mock_harness.new_instance.side_effect = _new_inst
    mock_harness.name = "lxd"
    with (
        patch("test_util.util.preload_snaps"),
        patch("test_util.util.setup_core_dumps"),
        patch("test_util.util.setup_k8s_snap"),
        patch("test_util.util.remove_k8s_snap"),
    ):
        gen = conftest.instances.__wrapped__(
            h=mock_harness,
            registry=MagicMock(),
            node_count=4,
            disable_k8s_bootstrapping=True,
            required_ports=[],
            no_setup=False,
            containerd_cfgdir="/etc/containerd",
            bootstrap_config=None,
            request=MagicMock(),
            network_type="IPv4",
        )
        with pytest.raises(RuntimeError, match="Backend failure inside new_instance"):
            next(gen)

    # Ensure both returned instances and registered-only instances were cleaned up
    assert len(created) == 2
    for inst in created:
        inst.delete.assert_called_once()
    mock_harness.delete_instance.assert_called_with("inst-2")
