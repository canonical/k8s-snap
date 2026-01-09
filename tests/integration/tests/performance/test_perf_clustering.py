#
# Copyright 2025 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import harness, tags, util


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.PERFORMANCE)
def test_perf_clustering_bootstrap_cli(instances: List[harness.Instance], benchmark):
    node = instances[0]

    def setup():
        # TODO(ben): benchmark `teardown` function is implemented but not yet released
        # in 5.1.0. Once released, we can move this teardown logic in a separate function.
        # See https://github.com/ionelmc/pytest-benchmark/issues/270
        node.exec(["snap", "remove", "k8s", "--purge"])
        util.setup_k8s_snap(node)

    def run():
        node.exec(["k8s", "bootstrap"])

    benchmark.pedantic(run, setup=setup, rounds=20)
