#
# Copyright 2024 Canonical, Ltd.
#
import logging
import time
from typing import List

import pytest
import yaml
from test_util import harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
def test_control_plane_taints(instances: List[harness.Instance]):
    k8s_instance = instances[0]

    bootstrap_conf = yaml.safe_dump(
        {"control-plane-taints": ["node-role.kubernetes.io/control-plane:NoSchedule"]}
    )

    k8s_instance.exec(
        ["dd", "of=/root/config.yaml"],
        input=str.encode(bootstrap_conf),
    )

    k8s_instance.exec(["k8s", "bootstrap", "--file", "/root/config.yaml"])
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
