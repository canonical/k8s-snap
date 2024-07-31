#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import Generator, List, Union

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

    _harness_clean(h)


def pytest_configure(config):
    config.addinivalue_line(
        "markers",
        "bootstrap_config: Provide a custom bootstrap config to the bootstrapping node.\n"
        "disable_k8s_bootstrapping: By default, the first k8s node is bootstrapped. This marker disables that.\n"
        "dualstack: Support dualstack on the instances.\n"
        "etcd_count: Mark a test to specify how many etcd instance nodes need to be created (None by default)\n"
        "node_count: Mark a test to specify how many instance nodes need to be created\n",
    )


@pytest.fixture(scope="function")
def node_count(request) -> int:
    node_count_marker = request.node.get_closest_marker("node_count")
    if not node_count_marker:
        return 1
    node_count_arg, *_ = node_count_marker.args
    return int(node_count_arg)


@pytest.fixture(scope="function")
def disable_k8s_bootstrapping(request) -> bool:
    return bool(request.node.get_closest_marker("disable_k8s_bootstrapping"))


@pytest.fixture(scope="function")
def bootstrap_config(request) -> Union[str, None]:
    bootstrap_config_marker = request.node.get_closest_marker("bootstrap_config")
    if not bootstrap_config_marker:
        return None
    config, *_ = bootstrap_config_marker.args
    return config


@pytest.fixture(scope="function")
def dualstack(request) -> bool:
    return bool(request.node.get_closest_marker("dualstack"))


@pytest.fixture(scope="function")
def instances(
    h: harness.Harness,
    node_count: int,
    tmp_path: Path,
    disable_k8s_bootstrapping: bool,
    bootstrap_config: str | None,
    dualstack: bool,
) -> Generator[List[harness.Instance], None, None]:
    """Construct instances for a cluster.

    Bootstrap and setup networking on the first instance, if `disable_k8s_bootstrapping` marker is not set.
    """
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    if node_count <= 0:
        pytest.xfail("Test requested 0 or fewer instances, skip this test.")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info(f"Creating {node_count} instances")
    instances: List[harness.Instance] = []

    for _ in range(node_count):
        # Create <node_count> instances and setup the k8s snap in each.
        instance = h.new_instance(dualstack=dualstack)
        instances.append(instance)
        util.setup_k8s_snap(instance, snap_path)

    if not disable_k8s_bootstrapping:
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
    # remove the session_instance. The harness ensures that everything is cleaned up
    # at the end of the test session.
    for instance in instances:
        h.delete_instance(instance.id)


@pytest.fixture(scope="session")
def session_instance(
    h: harness.Harness, tmp_path_factory: pytest.TempPathFactory
) -> Generator[harness.Instance, None, None]:
    """Constructs and bootstraps an instance that persists over a test session.

    Bootstraps the instance with all k8sd features enabled to reduce testing time.
    """
    LOG.info("Setup node and enable all features")

    snap_path = str(tmp_path_factory.mktemp("data") / "k8s.snap")
    instance = h.new_instance()
    util.setup_k8s_snap(instance, snap_path)

    bootstrap_config_path = "/home/ubuntu/bootstrap-session.yaml"
    instance.send_file(
        (config.MANIFESTS_DIR / "bootstrap-session.yaml").as_posix(),
        bootstrap_config_path,
    )

    instance.exec(["k8s", "bootstrap", "--file", bootstrap_config_path])
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    yield instance


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
