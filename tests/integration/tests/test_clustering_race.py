#
# Copyright 2026 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import harness, tags, util


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_wrong_token_race(instances: List[harness.Instance]):
    cluster_node = instances[0]

    join_token = util.get_join_token(cluster_node, instances[1])
    util.join_cluster(instances[1], join_token)

    new_join_token = util.get_join_token(cluster_node, instances[2])

    util.wait_until_k8s_ready(cluster_node, instances[:2])

    # retry since the truststore entry can be populated to cluster_node before
    # removing the node, otherwise the node removal will fail with
    # "No truststore entry found for node"
    # The heartbeat is every 2 seconds, so waiting for 3 seconds
    # should be sufficient.
    util.remove_node_with_retry(cluster_node, instances[1].id, retries=3)

    another_join_token = util.get_join_token(cluster_node, instances[2])

    # The join token should have changed after the node was removed as
    # it contains the ip addresses of all cluster nodes.
    assert (
        new_join_token != another_join_token
    ), "join token is not updated after node removal"
    util.join_cluster(instances[2], new_join_token)
