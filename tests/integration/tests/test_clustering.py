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


@pytest.mark.node_count(2)
@pytest.mark.snap_versions([util.previous_track(config.SNAP), config.SNAP])
def test_mixed_version_join(instances: List[harness.Instance]):
    """Test n versioned node joining a n-1 versioned cluster."""
    cluster_node = instances[0]  # bootstrapped on the previous channel
    joining_node = instances[1]  # installed with the snap under test

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

    # TODO: k8sd sometimes fails when requested to remove nodes immediately
    # after bootstrapping the cluster. It seems that it takes a little
    # longer for trust store changes to be propagated to all nodes, which
    # should probably be fixed on the microcluster side.
    #
    # For now, we'll perform some retries.
    #
    #   failed to POST /k8sd/cluster/remove: failed to delete cluster member
    #   k8s-integration-c1aee0-2: No truststore entry found for node with name
    #   "k8s-integration-c1aee0-2"
    util.stubbornly(retries=3, delay_s=5).on(cluster_node).exec(
        ["k8s", "remove-node", joining_cp.id]
    )
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "cp node should not have been removed from cluster"
    cluster_node.exec(["k8s", "remove-node", joining_worker.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "worker node should not have been removed from cluster"


@pytest.mark.node_count(3)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-skip-service-stop.yaml").read_text()
)
def test_skip_services_stop_on_remove(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    worker = instances[2]

    join_token = util.get_join_token(cluster_node, joining_cp)
    util.join_cluster(joining_cp, join_token)

    join_token_worker = util.get_join_token(cluster_node, worker, "--worker")
    util.join_cluster(worker, join_token_worker)

    util.wait_until_k8s_ready(cluster_node, instances)

    # TODO: skip retrying this once the microcluster trust store issue is addressed.
    util.stubbornly(retries=3, delay_s=5).on(cluster_node).exec(
        ["k8s", "remove-node", joining_cp.id]
    )
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "cp node should have been removed from the cluster"
    services = joining_cp.exec(
        ["snap", "services", "k8s"], capture_output=True, text=True
    ).stdout.split("\n")[1:-1]
    print(services)
    for service in services:
        if "k8s-apiserver-proxy" in service:
            assert (
                " inactive " in service
            ), "apiserver proxy should be inactive on control-plane"
        else:
            assert " active " in service, "service should be active"

    cluster_node.exec(["k8s", "remove-node", worker.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "worker node should have been removed from the cluster"
    services = worker.exec(
        ["snap", "services", "k8s"], capture_output=True, text=True
    ).stdout.split("\n")[1:-1]
    print(services)
    for service in services:
        for expected_active_service in [
            "containerd",
            "k8sd",
            "kubelet",
            "kube-proxy",
            "k8s-apiserver-proxy",
        ]:
            if expected_active_service in service:
                assert (
                    " active " in service
                ), f"{expected_active_service} should be active on worker"


@pytest.mark.node_count(3)
def test_join_with_custom_token_name(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp = instances[1]
    joining_cp_with_hostname = instances[2]

    out = cluster_node.exec(
        ["k8s", "get-join-token", "my-token"],
        capture_output=True,
        text=True,
    )
    join_token = out.stdout.strip()

    join_config = """
extra-sans:
- my-token
"""
    joining_cp.exec(
        ["k8s", "join-cluster", join_token, "--name", "my-node", "--file", "-"],
        input=join_config,
        text=True,
    )

    out = cluster_node.exec(
        ["k8s", "get-join-token", "my-token-2"],
        capture_output=True,
        text=True,
    )
    join_token_2 = out.stdout.strip()

    join_config_2 = """
extra-sans:
- my-token-2
"""
    joining_cp_with_hostname.exec(
        ["k8s", "join-cluster", join_token_2, "--file", "-"],
        input=join_config_2,
        text=True,
    )

    util.wait_until_k8s_ready(
        cluster_node, instances, node_names={joining_cp.id: "my-node"}
    )
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "nodes should have joined cluster"

    cluster_node.exec(["k8s", "remove-node", "my-node"])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "cp node should be removed from the cluster"

    cluster_node.exec(["k8s", "remove-node", joining_cp_with_hostname.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "cp node with hostname should be removed from the cluster"
