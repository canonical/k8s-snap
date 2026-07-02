#
# Copyright 2026 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster:\s*ready \(high-availability\)",
    r"nodes:\s*\d+\s*control-plane",
    r"^\s*$",
    r"Networking:",
    r"●\s*network\s*\(cilium",
    r"\s+\S.*",
    r"^\s*$",
    r"●\s*dns\s*\(coredns",
    r"\s+\S.*",
    r"^\s*$",
    r"○\s*load-balancer",
    r"○\s*ingress",
    r"●\s*gateway\s*\(cilium",
    r"\s+enabled",
    r"^\s*$",
    r"Storage:",
    r"●\s*local-storage\s*\(rawfile-csi",
    r"\s+\S.*",
    r"^\s*$",
    r"Observability:",
    r"●\s*metrics-server\s*\(metrics-server",
    r"\s+\S.*",
    r"^\s*$",
    r"Suggestions:",
    r"k8s kubectl get nodes\s+View detailed node information",
    r"k8s get\s+View cluster configuration",
]


@pytest.mark.tags(tags.WEEKLY)
@pytest.mark.node_count(3)
def test_restart(instances: List[harness.Instance]):
    """
    Test that a restart of the instance does not break the k8s snap.
    """

    main = instances[0]
    for joining in instances[1:]:
        token = util.get_join_token(main, joining)
        util.join_cluster(joining, token)

    LOG.info("Waiting for k8s to be ready")
    util.wait_until_k8s_ready(main, instances)
    for instance in instances:
        LOG.info("Waiting for the instance %s to be ready", instance.id)
        util.stubbornly(retries=15, delay_s=10).on(instance).until(
            condition=lambda p: util.status_output_matches(p, STATUS_PATTERNS),
        ).exec(["k8s", "status", "--wait-ready"])
        5

    # Verify all nodes are ready, not just k8s status
    util.wait_until_k8s_ready(main, instances)

    for instance in instances:
        LOG.info("Restart the instance %s", instance.id)
        instance.restart()

    LOG.info("Waiting for k8s to be ready after restart")
    util.wait_until_k8s_ready(instance, [instance])
    for instance in instances:
        LOG.info("Waiting for the instance %s to come back up", instance.id)
        util.stubbornly(retries=15, delay_s=10).on(instance).until(
            condition=lambda p: util.status_output_matches(p, STATUS_PATTERNS),
        ).exec(["k8s", "status", "--wait-ready"])

    # Verify all nodes are ready in Kubernetes after restart
    util.wait_until_k8s_ready(main, instances)
