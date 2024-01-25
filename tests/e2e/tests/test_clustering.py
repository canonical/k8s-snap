#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import List

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)


# Create <num_instances> instances and setup the k8s snap in each.
def setup_k8s_instances(
    h: harness.Harness, snap_path: str, num_instances: int
) -> List[str]:
    instances = []

    for _ in range(num_instances):
        instance_id = h.new_instance()
        instances.append(instance_id)
        util.setup_k8s_snap(h, instance_id, snap_path)

    return instances


# Create a token to join a node to an existing cluster
def add_node(h: harness.Harness, cluster_node: str, joining_node: str) -> str:
    out = h.exec(
        cluster_node,
        ["k8s", "add-node", joining_node],
        capture_output=True,
    )
    return out.stdout.decode().strip()


# Join an existing cluster.
def join_cluster(h: harness.Harness, instance_id, token):
    h.exec(instance_id, ["k8s", "join-cluster", token])


def test_clustering(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()
    instances = setup_k8s_instances(h, snap_path, num_instances=2)
    cluster_node = instances[0]
    joining_node = instances[1]

    h.exec(cluster_node, ["k8s", "bootstrap"])
    util.setup_network(h, cluster_node)

    h.exec(cluster_node, ["k8s", "kubectl", "get", "nodes", "-A"])

    token = add_node(h, cluster_node, joining_node)
    join_cluster(h, joining_node, token)

    util.wait_until_k8s_ready(h, cluster_node, instances)

    # TODO: Remove if --wait-ready for `join-cluster` is implemented.
    h.exec(cluster_node, ["k8s", "remove-node", util.hostname(h, joining_node)])

    h.cleanup()
