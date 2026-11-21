#
# Copyright 2026 Canonical, Ltd.
#
import threading
from concurrent.futures import ThreadPoolExecutor
from unittest.mock import MagicMock, patch

import pytest
from test_util import tags
from test_util.harness import HarnessError, Instance
from test_util.harness.bare_metal import BareMetalHarness
from test_util.harness.base import Harness
from test_util.harness.juju import JujuHarness
from test_util.harness.lxd import LXDHarness
from test_util.harness.multipass import MultipassHarness


@pytest.fixture(autouse=True)
def disable_inspection_reports():
    with patch("conftest.config.INSPECTION_REPORTS_DIR", None):
        yield


@pytest.mark.tags(tags.PULL_REQUEST)
def test_harness_has_lock():
    h = Harness()
    assert hasattr(h, "_lock")
    assert isinstance(h._lock, type(threading.Lock()))


@pytest.mark.tags(tags.PULL_REQUEST)
def test_lxd_and_multipass_next_id_thread_safety():
    with (
        patch.object(LXDHarness, "_configure_profile"),
        patch.object(LXDHarness, "_configure_network"),
    ):
        lxd_h = LXDHarness()
    assert hasattr(lxd_h, "_lock")

    mp_h = MultipassHarness.__new__(MultipassHarness)
    Harness.__init__(mp_h)
    mp_h._next_id = 0
    mp_h.instances = set()
    assert hasattr(mp_h, "_lock")

    def collect_ids(h, count):
        return [h.next_id() for _ in range(count)]

    for harness_obj in (lxd_h, mp_h):
        results = []
        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(collect_ids, harness_obj, 5) for _ in range(10)]
            for f in futures:
                results.extend(f.result())
        assert len(results) == 50
        assert len(set(results)) == 50
        assert sorted(results) == list(range(1, 51))


@pytest.mark.tags(tags.PULL_REQUEST)
def test_juju_harness_new_instance_thread_safety():
    with (
        patch("test_util.harness.juju.config.JUJU_CONTROLLER", "ctrl"),
        patch("test_util.harness.juju.config.JUJU_MODEL", "mdl"),
        patch("test_util.harness.juju.config.JUJU_MACHINES", "0 1 2 3"),
    ):
        h = JujuHarness()
    assert hasattr(h, "_lock")

    claimed_ids = []
    with ThreadPoolExecutor(max_workers=4) as executor:
        futures = [executor.submit(h.new_instance) for _ in range(4)]
        for f in futures:
            claimed_ids.append(f.result().id)

    assert sorted(claimed_ids) == ["0", "1", "2", "3"]
    assert h.instances == {"0", "1", "2", "3"}
    assert all(h.existing_machines.values())

    h.delete_instance("1")
    assert "1" not in h.instances
    assert not h.existing_machines["1"]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_harness_delete_instance_retains_id_on_failure():
    import subprocess

    lxd_h = LXDHarness.__new__(LXDHarness)
    Harness.__init__(lxd_h)
    lxd_h.instances = {"inst-0"}
    with patch(
        "test_util.harness.lxd.run",
        side_effect=subprocess.CalledProcessError(1, ["lxc", "rm"]),
    ):
        with pytest.raises(HarnessError):
            lxd_h.delete_instance("inst-0")
    assert "inst-0" in lxd_h.instances

    mp_h = MultipassHarness.__new__(MultipassHarness)
    Harness.__init__(mp_h)
    mp_h.instances = {"inst-0"}
    with patch(
        "test_util.harness.multipass.run",
        side_effect=subprocess.CalledProcessError(1, ["multipass"]),
    ):
        with pytest.raises(HarnessError):
            mp_h.delete_instance("inst-0")
    assert "inst-0" in mp_h.instances

    juju_h = JujuHarness.__new__(JujuHarness)
    Harness.__init__(juju_h)
    juju_h.instances = {"inst-0"}
    juju_h.existing_machines = {}
    with patch(
        "test_util.harness.juju.run",
        side_effect=subprocess.CalledProcessError(1, ["juju"]),
    ):
        with pytest.raises(HarnessError):
            juju_h.delete_instance("inst-0")
    assert "inst-0" in juju_h.instances


@pytest.mark.tags(tags.PULL_REQUEST)
def test_bare_metal_harness_single_instance_concurrent_rejection():
    h = BareMetalHarness.__new__(BareMetalHarness)
    Harness.__init__(h)
    h.instances = set()
    h._instance_id = "bare-metal-0"
    h._is_local = True
    h.exec = MagicMock()
    assert hasattr(h, "_lock")

    barrier = threading.Barrier(2)
    results = []
    errors = []

    def try_claim():
        barrier.wait()
        try:
            inst = h.new_instance()
            results.append(inst)
        except HarnessError as e:
            errors.append(e)

    t1 = threading.Thread(target=try_claim)
    t2 = threading.Thread(target=try_claim)
    t1.start()
    t2.start()
    t1.join()
    t2.join()

    assert len(results) == 1
    assert len(errors) == 1
    assert "bare_metal harness only supports a single instance" in str(errors[0])


@pytest.mark.tags(tags.PULL_REQUEST)
def test_instances_fixture_parallel_and_ordered():
    from conftest import instances as instances_fixture

    mock_h = MagicMock(spec=LXDHarness)
    mock_h.name = "lxd"
    mock_h.instances = set()

    import time

    active_threads = set()
    max_active_threads = 0
    lock = threading.Lock()

    created_instances = [Instance(mock_h, f"node-{i}") for i in range(4)]
    call_idx = 0

    def mock_new_instance(**kw):
        nonlocal call_idx, max_active_threads
        with lock:
            idx = call_idx
            call_idx += 1
            active_threads.add(threading.get_ident())
            if len(active_threads) > max_active_threads:
                max_active_threads = len(active_threads)
        # Delay earlier nodes so later nodes finish first if parallel
        time.sleep(0.04 if idx == 0 else 0.01)
        with lock:
            active_threads.discard(threading.get_ident())
        return created_instances[idx]

    mock_h.new_instance.side_effect = mock_new_instance
    mock_reg = MagicMock()
    mock_req = MagicMock()

    gen = instances_fixture.__wrapped__(
        h=mock_h,
        registry=mock_reg,
        node_count=4,
        disable_k8s_bootstrapping=True,
        required_ports=[],
        no_setup=True,
        containerd_cfgdir="/etc/containerd",
        bootstrap_config=None,
        request=mock_req,
        network_type="IPv4",
    )

    result_instances = next(gen)
    assert len(result_instances) == 4
    assert max_active_threads > 1
    for idx, inst in enumerate(result_instances):
        assert inst.id == f"node-{idx}"

    try:
        next(gen)
    except StopIteration:
        pass


@pytest.mark.tags(tags.PULL_REQUEST)
def test_instances_fixture_worker_capping():
    from conftest import instances as instances_fixture

    mock_h = MagicMock(spec=LXDHarness)
    mock_h.name = "lxd"
    mock_h.instances = set()
    mock_h.new_instance.side_effect = lambda **kw: Instance(mock_h, "node")
    mock_reg = MagicMock()
    mock_req = MagicMock()

    captured_max_workers = []

    real_init = ThreadPoolExecutor.__init__

    def spy_init(self, *args, **kwargs):
        captured_max_workers.append(kwargs.get("max_workers"))
        real_init(self, *args, **kwargs)

    with patch("os.cpu_count", return_value=4):
        with patch.object(ThreadPoolExecutor, "__init__", spy_init):
            gen = instances_fixture.__wrapped__(
                h=mock_h,
                registry=mock_reg,
                node_count=4,
                disable_k8s_bootstrapping=True,
                required_ports=[],
                no_setup=True,
                containerd_cfgdir="/etc/containerd",
                bootstrap_config=None,
                request=mock_req,
                network_type="IPv4",
            )
            next(gen)
            try:
                next(gen)
            except StopIteration:
                pass
    assert captured_max_workers[-1] == 2

    with patch("os.cpu_count", return_value=16):
        with patch.object(ThreadPoolExecutor, "__init__", spy_init):
            gen = instances_fixture.__wrapped__(
                h=mock_h,
                registry=mock_reg,
                node_count=4,
                disable_k8s_bootstrapping=True,
                required_ports=[],
                no_setup=True,
                containerd_cfgdir="/etc/containerd",
                bootstrap_config=None,
                request=mock_req,
                network_type="IPv4",
            )
            next(gen)
            try:
                next(gen)
            except StopIteration:
                pass
    assert captured_max_workers[-1] == 4


@pytest.mark.tags(tags.PULL_REQUEST)
def test_instances_fixture_cleanup_on_provisioning_failure():
    from conftest import instances as instances_fixture

    mock_h = MagicMock(spec=LXDHarness)
    mock_h.name = "lxd"
    mock_h.instances = set()

    inst0 = Instance(mock_h, "node-0")
    inst1 = Instance(mock_h, "node-1")

    def side_effect(**kw):
        if mock_h.new_instance.call_count == 1:
            mock_h.instances.add("node-0")
            return inst0
        elif mock_h.new_instance.call_count == 2:
            mock_h.instances.add("node-1")
            return inst1
        else:
            mock_h.instances.add("node-2-failed-seed")
            raise HarnessError("seed.loaded timed out")

    mock_h.new_instance.side_effect = side_effect
    mock_reg = MagicMock()
    mock_req = MagicMock()

    with (
        patch("conftest.util.setup_k8s_snap"),
        patch("conftest.util.remove_k8s_snap") as mock_remove,
    ):
        with pytest.raises(HarnessError, match="seed.loaded timed out"):
            gen = instances_fixture.__wrapped__(
                h=mock_h,
                registry=mock_reg,
                node_count=3,
                disable_k8s_bootstrapping=True,
                required_ports=[],
                no_setup=False,
                containerd_cfgdir="/etc/containerd",
                bootstrap_config=None,
                request=mock_req,
                network_type="IPv4",
            )
            next(gen)

        removed_instance_ids = [
            call_args[0][0].id for call_args in mock_remove.call_args_list
        ]
        assert "node-0" in removed_instance_ids
        assert "node-1" in removed_instance_ids
        assert "node-2-failed-seed" in removed_instance_ids
    deleted_instance_ids = [
        call_args[0][0] for call_args in mock_h.delete_instance.call_args_list
    ]
    assert "node-0" in deleted_instance_ids
    assert "node-1" in deleted_instance_ids
    assert "node-2-failed-seed" in deleted_instance_ids


@pytest.mark.tags(tags.PULL_REQUEST)
def test_instances_fixture_sequential_bare_metal_and_single_node():
    from conftest import instances as instances_fixture

    mock_h = MagicMock(spec=BareMetalHarness)
    mock_h.name = "bare_metal"
    mock_h.instances = set()
    mock_h.new_instance.return_value = Instance(mock_h, "bm-0")
    mock_reg = MagicMock()
    mock_req = MagicMock()

    with patch("conftest.ThreadPoolExecutor") as mock_tpe:
        gen = instances_fixture.__wrapped__(
            h=mock_h,
            registry=mock_reg,
            node_count=1,
            disable_k8s_bootstrapping=True,
            required_ports=[],
            no_setup=True,
            containerd_cfgdir="/etc/containerd",
            bootstrap_config=None,
            request=mock_req,
            network_type="IPv4",
        )
        res = next(gen)
        assert len(res) == 1
        assert not mock_tpe.called
        try:
            next(gen)
        except StopIteration:
            pass
