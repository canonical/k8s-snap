#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster status:\s*ready",
    r"control plane nodes:\s*(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})\s\(voter\)",
    r"high availability:\s*no",
    r"datastore:\s*etcd",
    r"network:\s*enabled",
    r"dns:\s*enabled at (\d{1,3}(?:\.\d{1,3}){3})",
    r"ingress:\s*disabled",
    r"load-balancer:\s*disabled",
    r"local-storage:\s*enabled at /var/snap/k8s/common/rawfile-storage",
    r"gateway\s*enabled",
]


@pytest.mark.tags(tags.WEEKLY)
@pytest.mark.node_count(3)
def test_reboot(instances: List[harness.Instance]):
    """
    Test that a reboot of the instance does not break the k8s snap.
    """

    for instance in instances:
        LOG.info("Waiting for the instance %s to be ready", instance.id)
        util.wait_until_k8s_ready(instance, [instance])
        util.stubbornly(retries=15, delay_s=10).on(instance).until(
            condition=lambda p: util.status_output_matches(p, STATUS_PATTERNS),
        ).exec(["k8s", "status", "--wait-ready"])

    for instance in instances:
        LOG.info("Rebooting the instance %s", instance.id)
        instance.reboot()

    for instance in instances:
        LOG.info("Waiting for the instance to come back up")
        util.wait_until_k8s_ready(instance, [instance])
        util.stubbornly(retries=15, delay_s=10).on(instance).until(
            condition=lambda p: util.status_output_matches(p, STATUS_PATTERNS),
        ).exec(["k8s", "status", "--wait-ready"])

    assert (
        len(util.ready_nodes(instances[0])) == 1
    ), "Expected exactly one ready node after reboot"
