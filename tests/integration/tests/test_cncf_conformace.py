#
# Copyright 2024 Canonical, Ltd.
#
import logging
import os
import platform
from typing import List

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)

ARCH = platform.machine()
ARCH_MAP = {"aarch64": "arm64", "x86_64": "amd64"}
SONOBUOY_VERSION = "v0.57.2"
SONOBUOY_TAR_GZ = f"https://github.com/vmware-tanzu/sonobuoy/releases/download/{SONOBUOY_VERSION}/sonobuoy_{SONOBUOY_VERSION[1: ]}_linux_{ARCH_MAP.get(ARCH)}.tar.gz"  # noqa


@pytest.mark.skipif(
    os.getenv("TEST_CNCF_E2E") in ["false", None],
    reason="Test is long and should be run nightly",
)
@pytest.mark.node_count(2)
def test_cncf_conformance(instances: List[harness.Instance]):
    cluster_node = cluster_setup(instances)
    install_sonobuoy(cluster_node)

    resp = cluster_node.exec(
        ["./sonobuoy"],
        capture_output=True,
    )
    assert resp.returncode == 0

    resp = cluster_node.exec(
        ["./sonobuoy", "run", "--plugin", "e2e", "--wait"],
        capture_output=True,
    )
    assert resp.returncode == 0

    resp = cluster_node.exec(
        ["./sonobuoy", "retrieve", "-f", "sonobuoy_e2e.tar.gz"],
        capture_output=True,
    )
    assert resp.returncode == 0

    resp = cluster_node.exec(
        ["tar", "-xf", "sonobuoy_e2e.tar.gz", "--one-top-level"],
        capture_output=True,
    )
    assert resp.returncode == 0

    resp = cluster_node.exec(
        ["./sonobuoy", "results", "sonobuoy_e2e.tar.gz"],
        capture_output=True,
        text=True,
    )
    assert resp.returncode == 0
    LOG.info(resp.stdout())
    pull_report(cluster_node)
    assert "Failed: 0" in resp.stdout()


def cluster_setup(instances: List[harness.Instance]) -> harness.Instance:
    cluster_node = instances[0]
    joining_node = instances[1]

    join_token = util.get_join_token(cluster_node, joining_node)
    util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 2, "node should have joined cluster"
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_node)

    # note: this workaround shlex issue when using string ">>"
    util.run(
        [
            "lxc",
            "shell",
            cluster_node.id,
            "--",
            "bash",
            "-c",
            "k8s config >> /root/.kube/config",
        ]
    )
    return cluster_node


def pull_report(instance: harness.Instance):
    instance.pull_file("/root/sonobuoy_e2e.tar.gz", "sonobuoy_e2e.tar.gz")


def install_sonobuoy(instance: harness.Instance):
    instance.exec(["curl", "-L", SONOBUOY_TAR_GZ, "-o", "sonobuoy.tar.gz"])
    instance.exec(["tar", "xvzf", "sonobuoy.tar.gz"])
    instance.exec(["./sonobuoy", "version"])
