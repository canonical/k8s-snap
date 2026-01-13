#
# Copyright 2026 Canonical, Ltd.
#
import logging
import re
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.CONFORMANCE)
def test_cncf_conformance(instances: List[harness.Instance]):
    cluster_node = cluster_setup(instances)
    install_sonobuoy(cluster_node)

    cluster_node.exec(
        [
            "./sonobuoy",
            "run",
            "--plugin",
            "e2e",
            "--mode",
            "certified-conformance",
            "--wait",
        ],
    )
    cluster_node.exec(
        ["./sonobuoy", "retrieve", "-f", "sonobuoy_e2e.tar.gz"],
    )
    cluster_node.exec(
        ["tar", "-xf", "sonobuoy_e2e.tar.gz", "--one-top-level"],
    )
    resp = cluster_node.exec(
        ["./sonobuoy", "results", "sonobuoy_e2e.tar.gz"],
        capture_output=True,
    )

    cluster_node.pull_file("/root/sonobuoy_e2e.tar.gz", "sonobuoy_e2e.tar.gz")

    output = resp.stdout.decode()
    LOG.info(output)
    failed_tests = int(re.search("Failed: (\\d+)", output).group(1))
    assert failed_tests == 0, f"{failed_tests} tests failed"


def cluster_setup(instances: List[harness.Instance]) -> harness.Instance:
    cluster_node = instances[0]
    joining_nodes = instances[1:]

    for joining_node in joining_nodes:
        join_token = util.get_join_token(cluster_node, joining_node)
        util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "node should have joined cluster"
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_nodes[0])
    assert "control-plane" in util.get_local_node_status(joining_nodes[1])

    config = cluster_node.exec(["k8s", "config"], capture_output=True)
    cluster_node.exec(["dd", "of=/root/.kube/config"], input=config.stdout)

    return cluster_node


def install_sonobuoy(instance: harness.Instance):
    instance.exec(
        ["curl", "-L", util.sonobuoy_tar_gz(instance.arch), "-o", "sonobuoy.tar.gz"]
    )
    instance.exec(["tar", "xvzf", "sonobuoy.tar.gz"])
    instance.exec(["./sonobuoy", "version"])
