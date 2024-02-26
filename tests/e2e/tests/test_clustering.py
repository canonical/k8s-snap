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

def reboot(instance):
    instance.exec(["reboot"])


@pytest.mark.node_count(2)
def test_clustering(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    token = add_node(cluster_node, joining_node)
    join_cluster(joining_node, token)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "node should have joined cluster"

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "node should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(2)
def test_worker_nodes(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_node = instances[1]

    token = add_node(cluster_node, joining_node, "--worker")
    join_cluster(joining_node, token)

    util.wait_until_k8s_ready(cluster_node, instances)
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "worker should have joined cluster"

    cluster_node.exec(["k8s", "remove-node", joining_node.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "worker should have been removed from cluster"
    assert (
        nodes[0]["metadata"]["name"] == cluster_node.id
    ), f"only {cluster_node.id} should be left in cluster"


@pytest.mark.node_count(1)
def test_instance_reboot_service_restart(instances: List[harness.Instance]):
    instance = instances[0]

    # Reboot the instance
    instance.reboot()

    # Wait until snap.k8s.kubelet service starts after the reboot
    util.stubbornly(retries=60, delay_s=5).on(lambda: instance.exec(
        ["systemctl", "is-active", "snap.k8s.kubelet"], capture_output=True
    )).until(lambda output: output.stdout.decode().strip() == "active")

    # Assert that the service is active after the retries
    service_status = instance.exec(["systemctl", "is-active", "snap.k8s.kubelet"], capture_output=True).stdout.decode().strip()
    assert service_status == "active", "snap.k8s.kubelet service did not start successfully"