#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(4)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-k8s-dqlite.yaml").read_text()
)
# For k8s-dqlite
@pytest.mark.required_ports(9000)
def test_control_plane_nodes_dqlite(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node_1 = instances[1]
    joining_node_2 = instances[2]
    joining_node_3 = instances[3]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_node_1)
    util.join_cluster(joining_node_1, join_token)

    join_token = util.get_join_token(cluster_node, joining_node_2)
    util.join_cluster(joining_node_2, join_token)

    join_token = util.get_join_token(cluster_node, joining_node_3)
    util.join_cluster(joining_node_3, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node_1)
    assert "control-plane" in util.get_local_node_status(joining_node_2)
    assert "control-plane" in util.get_local_node_status(joining_node_3)

    # Verify that the initial node can be removed
    joining_node_1.exec(["k8s", "remove-node", cluster_node.id])
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: not util.diverged_cluster_memberships(
            joining_node_1, [joining_node_1, joining_node_2, joining_node_3]
        )
    )

    # Verify that a node can remove itself
    joining_node_1.exec(["k8s", "remove-node", joining_node_1.id])
    util.stubbornly(retries=5, delay_s=3).until(
        lambda _: not util.diverged_cluster_memberships(
            joining_node_2, [joining_node_2, joining_node_3]
        )
    )


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-k8s-dqlite.yaml").read_text()
)
def test_worker_nodes_dqlite(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]
    other_joining_node = instances[2]

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_node, "--worker")
    join_token_2 = util.get_join_token(cluster_node, other_joining_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_node, join_token)

    util.join_cluster(other_joining_node, join_token_2)

    util.wait_until_k8s_ready(cluster_node, instances)

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
