#
# Copyright 2025 Canonical, Ltd.
#
import itertools
import logging
from pathlib import Path
from typing import Generator, Iterator, List, Optional, Union

import pytest
from test_util import config, harness, tags, util
from test_util.etcd import EtcdCluster
from test_util.registry import Registry

LOG = logging.getLogger(__name__)

pytest_plugins = ("pytest_tagging",)

# The following snaps will be downloaded once per test run and preloaded
# into the harness instances to reduce the number of downloads.
PRELOADED_SNAPS = ["snapd", "core20"]


def pytest_itemcollected(item):
    """
    A hook to ensure all tests have at least one tag before execution.
    """
    # Check for tags in the pytest.mark attributes
    marked_tags = [mark for mark in item.iter_markers(name="tags")]
    if not marked_tags or not any(
        tag.args[0] in tags.TEST_LEVELS for tag in marked_tags
    ):
        pytest.fail(
            f"The test {item.nodeid} does not have one of the test level tags."
            f"Please add at least one test-level tag using @pytest.mark.tags ({tags.TEST_LEVELS})."
        )


def _harness_clean(h: harness.Harness):
    "Clean up created instances within the test harness."

    if config.SKIP_CLEANUP:
        LOG.warning(
            "Skipping harness cleanup. "
            "It is your job now to clean up cloud resources"
        )
    else:
        LOG.debug("Cleanup")
        h.cleanup()


def _generate_inspection_report(h: harness.Harness, instance_id: str):
    LOG.debug("Generating inspection report for %s", instance_id)

    inspection_path = Path(config.INSPECTION_REPORTS_DIR)
    result = h.exec(
        instance_id,
        [
            "/snap/k8s/current/k8s/scripts/inspect.sh",
            "--all-namespaces",
            "--core-dump-dir",
            config.CORE_DUMP_DIR,
            "/inspection-report.tar.gz",
        ],
        capture_output=True,
        text=True,
        check=False,
    )

    (inspection_path / instance_id).mkdir(parents=True, exist_ok=True)
    report_log = inspection_path / instance_id / "inspection_report_logs.txt"
    with report_log.open("w") as f:
        f.write("stdout:\n")
        f.write(result.stdout)
        f.write("stderr:\n")
        f.write(result.stderr)

    try:
        h.pull_file(
            instance_id,
            "/inspection-report.tar.gz",
            (inspection_path / instance_id / "inspection_report.tar.gz").as_posix(),
        )
    except harness.HarnessError as e:
        LOG.warning("Failed to pull inspection report: %s", e)


@pytest.fixture(scope="session")
def h() -> harness.Harness:
    LOG.debug("Create harness for %s", config.SUBSTRATE)
    if config.SUBSTRATE == "lxd":
        h = harness.LXDHarness()
    elif config.SUBSTRATE == "multipass":
        h = harness.MultipassHarness()
    elif config.SUBSTRATE == "juju":
        h = harness.JujuHarness()
    else:
        raise harness.HarnessError(
            "TEST_SUBSTRATE must be one of: lxd, multipass, juju"
        )

    yield h

    if config.INSPECTION_REPORTS_DIR:
        for instance_id in h.instances:
            LOG.debug("Generating inspection reports for session instances")
            _generate_inspection_report(h, instance_id)

    _harness_clean(h)


@pytest.fixture(scope="session")
def registry(h: harness.Harness) -> Optional[Registry]:
    yield Registry(h)


@pytest.fixture(scope="session", autouse=True)
def snapd_preload() -> None:
    if not config.PRELOAD_SNAPS:
        LOG.info("Snap preloading disabled, skipping...")
        return

    LOG.info(f"Downloading snaps for preloading: {PRELOADED_SNAPS}")
    for snap in PRELOADED_SNAPS:
        util.run(
            [
                "snap",
                "download",
                snap,
                f"--basename={snap}",
                "--target-directory=/tmp",
            ]
        )


def pytest_configure(config):
    config.addinivalue_line(
        "markers",
        "bootstrap_config: Provide a custom bootstrap config to the bootstrapping node.\n"
        "disable_k8s_bootstrapping: By default, the first k8s node is bootstrapped. This marker disables that.\n"
        "no_setup: No setup steps (pushing snap, bootstrapping etc.) are performed on any node for this test.\n"
        "containerd_cfgdir: The instance containerd config directory, defaults to /etc/containerd."
        "network_type: Specify network type to use for the infrastructure (IPv4, Dualstack or IPv6).\n"
        "etcd_count: Mark a test to specify how many etcd instance nodes need to be created (None by default)\n"
        "node_count: Mark a test to specify how many instance nodes need to be created\n"
        "snap_versions: Mark a test to specify snap_versions for each node\n",
    )


@pytest.fixture(scope="function")
def node_count(request) -> int:
    node_count_marker = request.node.get_closest_marker("node_count")
    if not node_count_marker:
        return 1
    node_count_arg, *_ = node_count_marker.args
    return int(node_count_arg)


def snap_versions(request) -> Iterator[Optional[str]]:
    """An endless iterable of snap versions for each node in the test."""
    marking = ()
    if snap_version_marker := request.node.get_closest_marker("snap_versions"):
        marking, *_ = snap_version_marker.args
    # endlessly repeat of the configured snap version after exhausting the marking
    return itertools.chain(marking, itertools.repeat(None))


@pytest.fixture(scope="function")
def disable_k8s_bootstrapping(request) -> bool:
    return bool(request.node.get_closest_marker("disable_k8s_bootstrapping"))


@pytest.fixture(scope="function")
def no_setup(request) -> bool:
    return bool(request.node.get_closest_marker("no_setup"))


@pytest.fixture(scope="function")
def bootstrap_config(request) -> Union[str, None]:
    bootstrap_config_marker = request.node.get_closest_marker("bootstrap_config")
    if not bootstrap_config_marker:
        return None
    config, *_ = bootstrap_config_marker.args
    return config


@pytest.fixture(scope="function")
def network_type(request) -> Union[str, None]:
    bootstrap_config_marker = request.node.get_closest_marker("network_type")
    if not bootstrap_config_marker:
        return "IPv4"
    network_type, *_ = bootstrap_config_marker.args
    return network_type


@pytest.fixture(scope="function")
def containerd_cfgdir(request) -> str:
    marker = request.node.get_closest_marker("containerd_cfgdir")
    if not marker:
        return "/etc/containerd"
    cfgdir, *_ = marker.args
    return cfgdir


@pytest.fixture(scope="function")
def instances(
    h: harness.Harness,
    registry: Registry,
    node_count: int,
    tmp_path: Path,
    disable_k8s_bootstrapping: bool,
    no_setup: bool,
    containerd_cfgdir: str,
    bootstrap_config: Union[str, None],
    request,
    network_type: str,
) -> Generator[List[harness.Instance], None, None]:
    """Construct instances for a cluster.

    Bootstrap and setup networking on the first instance, if `disable_k8s_bootstrapping` marker is not set.
    """
    if node_count <= 0:
        pytest.xfail("Test requested 0 or fewer instances, skip this test.")

    LOG.info(f"Creating {node_count} instances")
    instances: List[harness.Instance] = []

    for _, snap in zip(range(node_count), snap_versions(request)):
        # Create <node_count> instances and setup the k8s snap in each.
        instance = h.new_instance(network_type=network_type)

        instance.exec("mkdir -p /etc/systemd/system/snapd.service.d".split())
        instance.exec("touch /etc/systemd/system/snapd.service.d/override.conf".split())
        # write file
        instance.exec(
            "echo -e '[Service]\nEnvironment=SNAPD_STANDBY_WAIT=1m' > /etc/systemd/system/snapd.service.d/override.conf".split()
        )
        instance.exec("systemctl daemon-reload".split())
        instance.exec("systemctl restart snapd.service".split())
        instances.append(instance)

        if config.PRELOAD_SNAPS:
            for preloaded_snap in PRELOADED_SNAPS:
                ack_file = f"{preloaded_snap}.assert"
                remote_path = (tmp_path / ack_file).as_posix()
                instance.send_file(
                    source=f"/tmp/{ack_file}",
                    destination=remote_path,
                )
                instance.exec(["snap", "ack", remote_path])

                snap_file = f"{preloaded_snap}.snap"
                remote_path = (tmp_path / snap_file).as_posix()
                instance.send_file(
                    source=f"/tmp/{snap_file}",
                    destination=remote_path,
                )
                instance.exec(["snap", "install", remote_path])

        if not no_setup:
            util.setup_core_dumps(instance)
            util.setup_k8s_snap(instance, tmp_path, snap)

            if config.USE_LOCAL_MIRROR:
                registry.apply_configuration(instance, containerd_cfgdir)

    if not disable_k8s_bootstrapping and not no_setup:
        first_node, *_ = instances

        if bootstrap_config:
            first_node.exec(
                ["k8s", "bootstrap", "--file", "-"],
                input=str.encode(bootstrap_config),
            )
        else:
            first_node.exec(["k8s", "bootstrap"])

    yield instances

    if config.SKIP_CLEANUP:
        LOG.warning("Skipping clean-up of instances, delete them on your own")
        return

    # Collect all the reports before initiating the cleanup so that we won't
    # affect the state of the observed cluster.
    if config.INSPECTION_REPORTS_DIR:
        for instance in instances:
            LOG.debug("Generating inspection reports for test instances")
            _generate_inspection_report(h, instance.id)

    # Cleanup after each test.
    # We cannot execute _harness_clean() here as this would also
    # remove session scoped instances. The harness ensures that everything is cleaned up
    # at the end of the test session.
    for instance in instances:
        try:
            util.remove_k8s_snap(instance)
        finally:
            h.delete_instance(instance.id)


@pytest.fixture(scope="function")
def etcd_count(request) -> int:
    etcd_count_marker = request.node.get_closest_marker("etcd_count")
    if not etcd_count_marker:
        return 0
    etcd_count_arg, *_ = etcd_count_marker.args
    return int(etcd_count_arg)


@pytest.fixture(scope="function")
def etcd_cluster(
    h: harness.Harness, etcd_count: int
) -> Generator[EtcdCluster, None, None]:
    """Construct etcd instances for a cluster."""
    LOG.info(f"Creating {etcd_count} etcd instances")

    cluster = EtcdCluster(h, initial_node_count=etcd_count)

    yield cluster
