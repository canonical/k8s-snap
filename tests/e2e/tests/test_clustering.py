#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from e2e_util import harness, util

LOG = logging.getLogger(__name__)


# Create a token to join a node to an existing cluster
def add_node(
    cluster_node: harness.Instance, joining_node: harness.Instance, *args: str
) -> str:
    out = cluster_node.exec(
        ["k8s", "add-node", joining_node.id, *args],
        capture_output=True,
    )
    return out.stdout.decode().strip()


# Join an existing cluster.
def join_cluster(instance, token):
    instance.exec(["k8s", "join-cluster", token])


@pytest.mark.node_count(2)
def test_clustering(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    token = add_node(cluster_node, joining_node)
    join_cluster(joining_node, token)

    util.wait_until_k8s_ready(cluster_node, instances)

    # TODO: Remove if --wait-ready for `join-cluster` is implemented.
    hostname = util.hostname(joining_node)
    util.stubbornly(retries=5, delay_s=3).on(cluster_node).exec(
        ["k8s", "remove-node", hostname]
    )


@pytest.mark.node_count(2)
def test_worker_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    token = add_node(cluster_node, joining_node, "--worker")
    join_cluster(joining_node, token)

    util.wait_until_k8s_ready(cluster_node, instances)
