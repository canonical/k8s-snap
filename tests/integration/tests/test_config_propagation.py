#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
def test_config_propagation(instances: List[harness.Instance]):
    initial_node = instances[0]
    joining_cplane_node = instances[1]
    joining_worker_node = instances[2]

    join_token = util.get_join_token(initial_node, joining_cplane_node)
    join_token_2 = util.get_join_token(initial_node, joining_worker_node, "--worker")

    assert join_token != join_token_2

    util.join_cluster(joining_cplane_node, join_token)

    util.join_cluster(joining_worker_node, join_token_2)

    util.wait_until_k8s_ready(initial_node, instances)
    nodes = util.ready_nodes(initial_node)
    assert len(nodes) == 3, "all nodes should have joined cluster"

    initial_node.exec(["k8s", "set", "dns.cluster-domain=integration.local"])

    util.stubbornly(retries=5, delay_s=10).on(joining_cplane_node).until(
        lambda p: f"--cluster-domain=integration.local" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])

    util.stubbornly(retries=5, delay_s=10).on(joining_worker_node).until(
        lambda p: f"--cluster-domain=integration.local" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])
