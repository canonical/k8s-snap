#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
from pathlib import Path
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

    # This is performed early since `k8s status` will fail after the node removal
    datastore_type = util.get_datastore_type(cluster_node)

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
    # We cannot determine the node type of the removed node, so we need to set it explicitly here.
    # NOTE: We're not expecting the k8sd service to be active after the node was removed.
    # microcluster removes the k8sd state folder, and without the "daemon.yaml" file in it,
    # k8sd fails to start.
    # NOTE: Etcd is skipped because it deactivates itself / exits
    # when the member remove API is called.
    util.check_snap_services_ready(
        joining_cp,
        node_type="control-plane",
        skip_services=["k8sd", "etcd"],
        datastore_type=datastore_type,
    )

    cluster_node.exec(["k8s", "remove-node", worker.id])
    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 1, "worker node should have been removed from the cluster"
    util.check_snap_services_ready(worker, node_type="worker", skip_services=["k8sd"])


@pytest.mark.node_count(2)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
# Old versions still use k8s-dqlite
@pytest.mark.required_ports(9000)
def test_disable_separate_feature_upgrades(
    instances: List[harness.Instance], tmp_path: Path, datastore_type: str
):
    cluster_node = instances[0]
    joining_cp = instances[1]

    start_branch = util.previous_track(config.SNAP)
    for instance in instances:
        instance.exec(f"snap install k8s --classic --channel={start_branch}".split())

    bootstrap_config = (
        config.MANIFESTS_DIR / "bootstrap-disable-separate-feature-upgrades.yaml"
    ).read_text()
    util.bootstrap(
        cluster_node, datastore_type=datastore_type, bootstrap_config=bootstrap_config
    )

    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    join_token = util.get_join_token(cluster_node, joining_cp)
    util.join_cluster(joining_cp, join_token)

    util.wait_until_k8s_ready(joining_cp, instances)

    # Refresh first node, no upgrade CRD should be created.
    util.setup_k8s_snap(cluster_node, config.SNAP)
    util.wait_until_k8s_ready(cluster_node, instances)

    upgrades = json.loads(
        cluster_node.exec(
            "k8s kubectl get upgrade -o=jsonpath={.items}".split(),
            capture_output=True,
            text=True,
        ).stdout
    )
    assert len(upgrades) == 0, "upgrade CRD should not be created"

    # The feature controller should not be blocked.
    # Disable gateway feature
    cluster_node.exec("k8s set gateway.enabled=false".split())

    def is_gateway_disabled(process):
        gateway_status = json.loads(process.stdout)
        return gateway_status.get("enabled") is False

    # Wait until gateway is disabled
    util.stubbornly(retries=3, delay_s=5).on(cluster_node).until(
        is_gateway_disabled
    ).exec(
        ["k8s", "get", "gateway", "--output-format=json"],
        text=True,
        capture_output=True,
    )
