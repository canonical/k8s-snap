#
# Copyright 2025 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import harness, tags, util


# Note(ben): Commented out as otherwise the setup would still happen for xfail tests.
# @pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_wrong_token_race(instances: List[harness.Instance]):
    # Note(ben): k8s-dqlite sometimes takes very long to shutdown (to be investigated) and
    # since microcluster has a 30s fixed timeout for the remove hooks this test sometimes fails.
    # The timeout will be configurable in https://github.com/canonical/microcluster/pull/365)
    # The k8s-dqlite issue will be investigated separately.
    pytest.xfail("This test is currently flaky because of a k8s-dqlite shutdown issue.")
    cluster_node = instances[0]

    join_token = util.get_join_token(cluster_node, instances[1])
    util.join_cluster(instances[1], join_token)

    new_join_token = util.get_join_token(cluster_node, instances[2])

    util.wait_until_k8s_ready(cluster_node, instances[:2])
    cluster_node.exec(["k8s", "remove-node", instances[1].id])

    another_join_token = util.get_join_token(cluster_node, instances[2])

    # The join token should have changed after the node was removed as
    # it contains the ip addresses of all cluster nodes.
    assert (
        new_join_token != another_join_token
    ), "join token is not updated after node removal"
    util.join_cluster(instances[2], new_join_token)
