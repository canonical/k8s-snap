#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)

CONTAINERD_PATHS = [
    "/etc/containerd",
    "/opt/cni/bin",
    "/run/containerd",
    "/var/lib/containerd",
]


@pytest.mark.node_count(1)
def test_node_cleanup(instances: List[harness.Instance], tmp_path):
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    util.remove_k8s_snap(instance)

    # Check that the containerd-related folders are removed on snap removal.
    process = instance.exec(
        ["ls", *CONTAINERD_PATHS], capture_output=True, text=True, check=False
    )
    for path in CONTAINERD_PATHS:
        assert f"cannot access '{path}': No such file or directory" in process.stderr
        
    util.setup_k8s_snap(instance, tmp_path)
    instance.exec(["k8s", "bootstrap"])

