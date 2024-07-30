#
# Copyright 2024 Canonical, Ltd.
#
import logging
import time
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    config.MANIFESTS_DIR / "bootstrap-control-plane-taints.yaml"
)
def test_control_plane_taints(instances: List[harness.Instance]):
    k8s_instance = instances[0]
    retries = 10

    while retries and not (nodes := util.get_nodes(k8s_instance)):
        LOG.info("Waiting for Nodes")
        time.sleep(3)
        retries -= 1
    assert len(nodes) == 1, "Should have found one node in 30 sec"
    assert all(
        [
            t["effect"] == "NoSchedule"
            for t in nodes[0]["spec"]["taints"]
            if t["key"] == "node-role.kubernetes.io/control-plane"
        ]
    )
