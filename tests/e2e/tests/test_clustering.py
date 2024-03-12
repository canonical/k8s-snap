#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
from typing import List

import pytest
from e2e_util import harness, util

LOG = logging.getLogger(__name__)


# Create a token to join a node to an existing cluster
def get_join_token(
    cluster_node: harness.Instance, joining_node: harness.Instance, *args: str
) -> str:
    out = cluster_node.exec(
        ["k8s", "get-join-token", joining_node.id, "--output-format", "json", *args],
        capture_output=True,
    )
    result = json.loads(out.stdout.decode())
    return result["join-token"]


# Join an existing cluster.
def join_cluster(instance, join_token):
    instance.exec(["k8s", "join-cluster", join_token])


@pytest.mark.node_count(2)
def test_clustering(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    join_token = get_join_token(cluster_node, joining_node)
    join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "node should have joined cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "node should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(3)
def test_worker_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]
    other_joining_node = instances[2]

    join_token = get_join_token(cluster_node, joining_node, "--worker")
    join_token_2 = get_join_token(cluster_node, other_joining_node, "--worker")

    assert join_token != join_token_2
    
    join_cluster(joining_node, join_token)
    
    join_cluster(other_joining_node, join_token_2)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "workers should have joined cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "worker" in util.get_local_node_status(joining_node)
    assert "worker" in util.get_local_node_status(other_joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "worker should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"
