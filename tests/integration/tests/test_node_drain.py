#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.node_count(3)
def test_node_drain(instances: List[harness.Instance]):
    """
    Test that the node is drained upon removal.
    """

    cp = instances[0]
    cp2 = instances[1]
    worker = instances[2]
    cp_token = util.get_join_token(cp, cp2)
    worker_token = util.get_join_token(cp, worker, "--worker")
    util.join_cluster(cp2, cp_token)
    util.join_cluster(worker, worker_token)

    util.wait_until_k8s_ready(cp, instances)

    util.stubbornly(retries=3, delay_s=5).on(cp).exec(["k8s", "remove-node", cp2.id])
    util.stubbornly(retries=3, delay_s=5).on(cp).exec(["k8s", "remove-node", worker.id])

    util.stubbornly(retries=3, delay_s=5).on(cp).until(
        lambda p: nodes_drained(cp, [cp2.id, worker.id])
    )


def nodes_drained(instance: harness.Instance, node_ids: List[str]) -> bool:
    for node_id in node_ids:
        pods = instance.exec(
            [
                "k8s",
                "kubectl",
                "get",
                "pods",
                "-A",
                f"--field-selector=spec.nodeName={node_id}",
                "-o",
                "jsonpath={.items..metadata.name",
            ],
            check=True,
            output=True,
        )
        if pods:
            LOG.info(f"Node {node_id} still has pods: {pods}")
            return False
    return True
