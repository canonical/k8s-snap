#
# Copyright 2026 Canonical, Ltd.
#
import logging
import time
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    'control-plane-taints: ["node-role.kubernetes.io/control-plane:NoSchedule"]'
)
@pytest.mark.tags(tags.NIGHTLY)
def test_control_plane_taints(instances: List[harness.Instance]):
    k8s_instance = instances[0]
    util.wait_until_k8s_ready(k8s_instance, [k8s_instance])

    retries = 10
    nodes = []
    while retries:
        nodes = util.get_nodes(k8s_instance)
        if nodes and "taints" in nodes[0].get("spec", {}):
            break
        LOG.info("Waiting for Nodes and taints")
        time.sleep(3)
        retries -= 1

    assert len(nodes) == 1, "Should have found one node in 30 sec"
    matching_taints = [
        t
        for t in nodes[0].get("spec", {}).get("taints", [])
        if t.get("key") == "node-role.kubernetes.io/control-plane"
    ]
    assert matching_taints, "Expected node-role.kubernetes.io/control-plane taint"
    assert all(t.get("effect") == "NoSchedule" for t in matching_taints)
