#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import Generator, List

import pytest
from test_util import config, harness, util

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
        "node_count: Mark a test to specify how many instance nodes need to be created",
    )


@pytest.fixture(scope="function")
def node_count(request) -> int:
    node_count_marker = request.node.get_closest_marker("node_count")
    if not node_count_marker:
        return 1
    node_count_arg, *_ = node_count_marker.args
    return int(node_count_arg)


@pytest.fixture(scope="function")
def instances(
    h: harness.Harness, node_count: int, tmp_path: Path
) -> Generator[List[harness.Instance], None, None]:
    """Construct instances for a cluster.

    Bootstrap and setup networking on the first instance.
    """
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    if node_count <= 0:
        pytest.xfail("Test requested 0 or fewer instances, skip this test.")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info(f"Creating {node_count} instances")
    instances: List[util.Instance] = []

    for _ in range(node_count):
        # Create <node_count> instances and setup the k8s snap in each.
        instance = h.new_instance()
        instances.append(instance)
        util.setup_k8s_snap(instance, snap_path)

    first_node, *_ = instances
    first_node.exec(["k8s", "bootstrap"])

    yield instances

    _harness_clean(h)
