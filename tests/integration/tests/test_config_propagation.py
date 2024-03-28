#
# Copyright 2024 Canonical, Ltd.
#
import json
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

    p = (
        util.stubbornly(retries=5, delay_s=3)
        .on(initial_node)
        .until(lambda p: len(p.stdout.decode().replace("'", "")) > 0)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "service",
                "coredns",
                "-n",
                "kube-system",
                "-o=jsonpath='{.spec.clusterIP}'",
            ],
        )
    )
    service_ip = p.stdout.decode().replace("'", "")

    util.stubbornly(retries=5, delay_s=10).on(joining_cplane_node).until(
        lambda p: f"--cluster-dns={service_ip}" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])

    util.stubbornly(retries=5, delay_s=10).on(joining_worker_node).until(
        lambda p: f"--cluster-dns={service_ip}" in p.stdout.decode()
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])
