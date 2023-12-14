#
# Copyright 2023 Canonical, Ltd.
#
import base64
import logging
from pathlib import Path
from typing import List

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)
# TODO(bschimke): Remove when snap supports CLI alias
K8S_BINARY_PATH = "/snap/k8s/current/bin/k8s"


# Create <num_instances> instances and setup the k8s snap in each.
def setup_k8s_instances(
    h: harness.Harness, snap_path: str, num_instances: int, wait_ready: bool = True
) -> List[str]:
    instances = []

    for _ in range(num_instances):
        instance_id = h.new_instance()
        instances.append(instance_id)
        util.setup_k8s_snap(h, instance_id, snap_path, wait_ready=wait_ready)

    return instances


# Bootstrap microcluster on this instance
def bootstrap_cluster(h: harness.Harness, instance_id: str):
    out = h.exec(
        instance_id,
        [K8S_BINARY_PATH, "bootstrap-cluster"],
        capture_output=True,
    )
    assert "created" in out.stderr.decode()


# Create a token to join a node to an existing cluster
def add_node(h: harness.Harness, cluster_node: str, joining_node: str) -> str:
    out = h.exec(
        cluster_node,
        [K8S_BINARY_PATH, "add-node", joining_node],
        capture_output=True,
    )
    token = out.stdout.decode().strip()
    assert (
        base64.b64encode(base64.b64decode(token)).decode() == token
    ), f"add-node should return a base64 token but got {token}"
    return token


# Join an existing cluster.
def join_cluster(h: harness.Harness, instance_id, token):
    out = h.exec(
        instance_id,
        [K8S_BINARY_PATH, "join-node", token],
        capture_output=True,
    )
    LOG.info(out.stdout.decode())
    LOG.info(out.stderr.decode())
    assert f"Joined {instance_id}" in out.stderr.decode()


def test_clustering(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()
    instances = setup_k8s_instances(h, snap_path, num_instances=2, wait_ready=False)
    cluster_node = instances[0]
    joining_node = instances[1]

    bootstrap_cluster(h, cluster_node)
    token = add_node(h, cluster_node, joining_node)
    join_cluster(h, joining_node, token)
