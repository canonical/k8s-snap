#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from e2e_util import harness

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
def test_kubeconfig(instances: List[harness.Instance]):
    instance = instances[0]
    result = instance.exec(
        "k8s config --server 192.168.210.41".split(), capture_output=True
    )
    config = result.stdout.decode()
    assert len(config) > 0
    assert "server: https://192.168.210.41" in config
