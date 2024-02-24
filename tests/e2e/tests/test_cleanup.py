#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from e2e_util import harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
def test_node_cleanup(instances: List[harness.Instance]):
    instance = instances[0]
    util.setup_dns(instance)

    LOG.info("Uninstall k8s...")
    instance.exec(["snap", "remove", "k8s", "--purge"])

    LOG.info("Waiting for shims to go away...")
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["containerd-shim", "cilium", "coredns", "/pause"]
        )
    ).exec(["ps", "-fea"])

    LOG.info("Waiting for CNI network namespaces to go away...")
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: "cni-" not in p.stdout.decode()
    ).exec(["ip", "netns", "list"])

    LOG.info("Waiting for kubelet and containerd mounts to go away...")
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["/var/lib/kubelet/pods", "/run/containerd/io.containerd"]
        )
    ).exec(["mount"])
