#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List
import time
import base64
import json

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)



@pytest.mark.node_count(3)
def test_wrong_token_race(instances: List[harness.Instance]):
    cluster_node = instances[0]

    join_token = util.get_join_token(cluster_node, instances[1])
    util.join_cluster(instances[1], join_token)

    new_join_token = util.get_join_token(cluster_node, instances[2])

    cluster_node.exec(["k8s", "remove-node", instances[1].id])

    another_join_token = util.get_join_token(cluster_node, instances[2])

    # The join token should have changed after the node was removed as it contains the ip addresses of all cluster nodes.
    assert new_join_token != another_join_token
    util.join_cluster(instances[2], new_join_token)
