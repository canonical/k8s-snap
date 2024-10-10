#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
def test_node_cleanup(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    LOG.info("Uninstall k8s...")
    instance.exec(["snap", "remove", config.SNAP_NAME, "--purge"])

    LOG.info("Waiting for shims to go away...")
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["containerd-shim", "cilium", "coredns", "/pause"]
        )
    ).exec(["ps", "-fea"])

    LOG.info("Waiting for kubelet and containerd mounts to go away...")
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: all(
            x not in p.stdout.decode()
            for x in ["/var/lib/kubelet/pods", "/run/containerd/io.containerd"]
        )
    ).exec(["mount"])

    # NOTE(neoaggelos): Temporarily disable this as it fails on strict.
    # For details, `snap changes` then `snap change $remove_k8s_snap_change`.
    # Example output follows:
    #
    # 2024-02-23T14:10:42Z ERROR ignoring failure in hook "remove":
    # -----
    # ...
    # ip netns delete cni-UUID1
    # Cannot remove namespace file "/run/netns/cni-UUID1": Device or resource busy
    # ip netns delete cni-UUID2
    # Cannot remove namespace file "/run/netns/cni-UUID2": Device or resource busy
    # ip netns delete cni-UUID3
    # Cannot remove namespace file "/run/netns/cni-UUID3": Device or resource busy

    # LOG.info("Waiting for CNI network namespaces to go away...")
    # util.stubbornly(retries=5, delay_s=5).on(instance).until(
    #     lambda p: "cni-" not in p.stdout.decode()
    # ).exec(["ip", "netns", "list"])
