#
# Copyright 2024 Canonical, Ltd.
#
import itertools
import logging
from pathlib import Path
from typing import Generator, Iterator, List, Optional, Union

import pytest
from test_util import config, harness, util
from test_util.etcd import EtcdCluster

LOG = logging.getLogger(__name__)


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
        ["/snap/k8s/current/k8s/scripts/inspect.sh", "/inspection-report.tar.gz"],
        capture_output=True,
        text=True,
        check=False,
    )

    (inspection_path / instance_id).mkdir(parents=True, exist_ok=True)
    (inspection_path / instance_id / "inspection_report_logs.txt").write_text(
        result.stdout
    )

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
    if config.SUBSTRATE == "local":
        h = harness.LocalHarness()
    elif config.SUBSTRATE == "lxd":
        h = harness.LXDHarness()
    elif config.SUBSTRATE == "multipass":
        h = harness.MultipassHarness()
    elif config.SUBSTRATE == "juju":
        h = harness.JujuHarness()
    else:
        raise harness.HarnessError(
            "TEST_SUBSTRATE must be one of: local, lxd, multipass, juju"
        )

    yield h

    if config.INSPECTION_REPORTS_DIR is not None:
        for instance_id in h.instances:
            LOG.debug("Generating inspection reports for session instances")
            _generate_inspection_report(h, instance_id)

    _harness_clean(h)


def pytest_configure(config):
    config.addinivalue_line(
        "markers",
        "bootstrap_config: Provide a custom bootstrap config to the bootstrapping node.\n"
        "disable_k8s_bootstrapping: By default, the first k8s node is bootstrapped. This marker disables that.\n"
        "no_setup: No setup steps (pushing snap, bootstrapping etc.) are performed on any node for this test.\n"
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
def instances(
    h: harness.Harness,
    node_count: int,
    tmp_path: Path,
    disable_k8s_bootstrapping: bool,
    no_setup: bool,
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
        instances.append(instance)
        if not no_setup:
            util.setup_k8s_snap(instance, tmp_path, snap)

    if not disable_k8s_bootstrapping and not no_setup:
        first_node, *_ = instances

        if bootstrap_config is not None:
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

    # Cleanup after each test.
    # We cannot execute _harness_clean() here as this would also
    # remove session scoped instances. The harness ensures that everything is cleaned up
    # at the end of the test session.
    for instance in instances:
        if config.INSPECTION_REPORTS_DIR is not None:
            LOG.debug("Generating inspection reports for test instances")
            _generate_inspection_report(h, instance.id)

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
