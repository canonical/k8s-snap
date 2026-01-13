#
# Copyright 2026 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import harness, tags


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.PERFORMANCE)
def test_perf_status_single_node_cli(instances: List[harness.Instance], benchmark):
    node = instances[0]

    def run():
        node.exec(["k8s", "status"])

    benchmark.pedantic(run, rounds=40)
