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



@pytest.mark.node_count(5)
def test_wrong_token_race(instances: List[harness.Instance]):
    cluster_node = instances[0]

    for instance in instances[1:]:
        join_token = util.get_join_token(cluster_node, instance)
        util.join_cluster(instance, join_token)

    cluster_node.exec(["k8s", "remove-node", instances[1].id])
    new_join_token = util.get_join_token(cluster_node, instances[1])
    time.sleep(20)
    another_join_token = util.get_join_token(cluster_node, instances[1])

    print(f"new_join_token: {json.dumps(json.loads(base64.b64decode(new_join_token)), indent=2)}")
    print({json.dumps(json.loads(base64.b64decode(new_join_token)), indent=2)})
    print("-" * 80)
    print(f"another_join_token: {another_join_token}")
    print(f"another_join_token: {json.dumps(json.loads(base64.b64decode(another_join_token)), indent=2)}")
    assert new_join_token == another_join_token, "Tokens should be the same"
