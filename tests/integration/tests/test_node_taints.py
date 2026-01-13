#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from typing import Any, List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-node-taints.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_node_taints(instances: List[harness.Instance]):
    """Test node taints are returned by the GetNodeStatus RPC."""
    cp_node = instances[0]
    worker_node = instances[1]
    join_token = util.get_join_token(cp_node, worker_node, "--worker")
    util.join_cluster(
        worker_node,
        join_token,
        (config.MANIFESTS_DIR / "worker-join-node-taints.yaml").read_text(),
    )

    util.wait_until_k8s_ready(cp_node, instances)
    nodes = util.ready_nodes(cp_node)
    assert len(nodes) == 2, "worker should have joined cluster"

    cp_taints = get_node_taints(cp_node)
    worker_taints = get_node_taints(worker_node)

    # NOTE(Hue): these come from the bootstrap and join configs
    cp_exp_taints = [
        "taint1=:PreferNoSchedule",
        "taint2=value:PreferNoSchedule",
    ]
    worker_exp_taints = [
        "workerTaint1=:PreferNoSchedule",
        "workerTaint2=workerValue:PreferNoSchedule",
    ]

    assert len(cp_taints) == len(cp_exp_taints), "cp node taints count do not match"
    assert len(worker_taints) == len(
        worker_exp_taints
    ), "worker node taints count do not match"
    assert set(cp_taints) == set(cp_exp_taints), "cp node taints do not match"
    assert set(worker_taints) == set(
        worker_exp_taints
    ), "worker node taints do not match"


def get_node_taints(instance: harness.Instance) -> Any:
    """Get taints from the node status."""
    resp = instance.exec(
        [
            "curl",
            "-H",
            "Content-Type: application/json",
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/k8sd/node",
        ],
        capture_output=True,
    )
    assert resp.returncode == 0, "Failed to get node status."
    response = json.loads(resp.stdout.decode())
    assert response["error_code"] == 0, "Failed to get node status."
    assert response["error"] == "", "Failed to get node status."

    metadata = response.get("metadata")
    assert metadata is not None, "Metadata not found in the node status response."
    taints = metadata.get("taints")
    assert taints is not None, "Node taints not found in the node status response."

    return taints
