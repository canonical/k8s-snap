#
# Copyright 2025 Canonical, Ltd.
#
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.PERFORMANCE)
def test_perf_clustering_bootstrap_cli(
    instances: List[harness.Instance], benchmark, datastore_type: str
):
    node = instances[0]
    bootstrap_yaml_bytes = None

    def setup():
        nonlocal bootstrap_yaml_bytes
        # TODO(ben): benchmark `teardown` function is implemented but not yet released
        # in 5.1.0. Once released, we can move this teardown logic in a separate function.
        # See https://github.com/ionelmc/pytest-benchmark/issues/270
        node.exec(["snap", "remove", "k8s", "--purge"])
        util.setup_k8s_snap(node)

        # Prepare bootstrap configuration (move YAML operations out of benchmark)
        default_config_path = config.MANIFESTS_DIR / "bootstrap-default.yaml"
        bootstrap_config = yaml.safe_load(default_config_path.read_text())
        bootstrap_config["datastore-type"] = datastore_type
        bootstrap_yaml = yaml.dump(bootstrap_config, default_flow_style=False)
        bootstrap_yaml_bytes = str.encode(bootstrap_yaml)

    def run():
        # Only run the actual bootstrap command
        node.exec(
            ["k8s", "bootstrap", "--file", "-"],
            input=bootstrap_yaml_bytes,
        )

    benchmark.pedantic(run, setup=setup, rounds=20)
