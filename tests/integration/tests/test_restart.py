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
    r"high availability:\s*yes",
    r"network:\s*enabled",
    r"dns:\s*enabled at (\d{1,3}(?:\.\d{1,3}){3})",
    r"ingress:\s*disabled",
    r"load-balancer:\s*disabled",
    r"local-storage:\s*enabled at /var/snap/k8s/common/rawfile-storage",
    r"gateway\s*enabled",
]


@pytest.mark.tags(tags.WEEKLY)
@pytest.mark.node_count(3)
def test_restart(instances: List[harness.Instance], datastore_type: str):
    """
    Test that a restart of the instance does not break the k8s snap.
    """
    STATUS_PATTERNS.insert(3, r"datastore:\s*{}".format(datastore_type))

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

    # An additional check to ensure the cluster is still functional
    assert len(util.ready_nodes(main)) == 3
