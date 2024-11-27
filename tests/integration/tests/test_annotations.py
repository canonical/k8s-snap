#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-no-k8s-node-remove.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
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
@pytest.mark.tags(tags.NIGHTLY)
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
