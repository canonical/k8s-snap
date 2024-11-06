#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
def test_node_cleanup(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    util.remove_k8s_snap(instance)
