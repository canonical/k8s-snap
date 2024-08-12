#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
def test_control_plane_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    join_token = util.get_join_token(cluster_node, joining_node)
    util.join_cluster(joining_node, join_token)

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

    join_token = util.get_join_token(cluster_node, joining_node, "--worker")
    join_token_2 = util.get_join_token(cluster_node, other_joining_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_node, join_token)

    util.join_cluster(other_joining_node, join_token_2)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "workers should have joined cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "worker" in util.get_local_node_status(joining_node)
    assert "worker" in util.get_local_node_status(other_joining_node)

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "worker should have been removed from cluster"
    assert cluster_node.id in [
        node["metadata"]["name"] for node in nodes
    ] and other_joining_node.id in [
        node["metadata"]["name"] for node in nodes
    ], f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(3)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-no-k8s-node-remove.yaml").read_text()
)
def test_no_remove(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    joining_worker = instances[2]

    join_token = util.get_join_token(cluster_node, joining_cp)
    join_token_worker = util.get_join_token(cluster_node, joining_worker, "--worker")
    util.join_cluster(joining_cp, join_token)
    util.join_cluster(joining_worker, join_token_worker)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "nodes should have joined cluster"

    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_cp)
    assert "worker" in util.get_local_node_status(joining_worker)

    nodes = util.ready_nodes(cluster_node)

    cluster_node.exec(["k8s", "remove-node", joining_cp.id])
    assert len(nodes) == 3, "cp node should not have been removed from cluster"
    cluster_node.exec(["k8s", "remove-node", joining_worker.id])
    assert len(nodes) == 3, "worker node should not have been removed from cluster"


@pytest.mark.node_count(5)
def test_replacing_control_plane_nodes(instances: List[harness.Instance]):
    cp1 = instances[0]
    cp2 = instances[1]
    cp3 = instances[2]
    new_cp4 = instances[3]
    new_cp5 = instances[4]

    # initial 3 node setup
    join_token_2 = util.get_join_token(cp1, cp2)
    util.join_cluster(cp2, join_token_2)
    join_token_3 = util.get_join_token(cp1, cp3)   
    util.join_cluster(cp3, join_token_3)
    util.wait_until_k8s_ready(cp1, instances[:3])
    nodes = util.ready_nodes(cp1)
    assert len(nodes) == 3, f"initial nodes should have joined the cluster, expected 3, got {len(nodes)}"

    # adding the first new node
    join_token_4 = util.get_join_token(cp1, new_cp4)
    util.join_cluster(new_cp4, join_token_4)
    util.wait_until_k8s_ready(cp1, instances[:4])
    nodes = util.ready_nodes(cp1)
    assert len(nodes) == 4, f"first new node should have joined the cluster, expected 4, got {len(nodes)}"

    # removing the first old node
    cp3.exec(["k8s", "remove-node", cp1.id])
    util.wait_until_k8s_ready(cp3, instances[1:4])
    nodes = util.ready_nodes(cp3)
    assert len(nodes) == 3, f"first old node should have been removed from the cluster, expected 3, got {len(nodes)}"

    # adding the second new node
    join_token_5 = util.get_join_token(cp3, new_cp5)
    util.join_cluster(new_cp5, join_token_5)
    util.wait_until_k8s_ready(cp3, instances[1:5])
    nodes = util.ready_nodes(cp3)
    assert len(nodes) == 4, f"second new node should have joined cluster, expected 4, got {len(nodes)}"

    # removing the second old node
    cp3.exec(["k8s", "remove-node", cp2.id])
    util.wait_until_k8s_ready(cp3, instances[2:5])
    nodes = util.ready_nodes(cp3)
    assert len(nodes) == 3, f"second old node should have been removed from the cluster, expected 3, got {len(nodes)}"
