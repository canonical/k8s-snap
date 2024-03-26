#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

from e2e_util import harness, util

LOG = logging.getLogger(__name__)


def test_smoke(instances: List[harness.Instance]):
    instance = instances[0]

    util.wait_until_k8s_ready(instance, instances)

    # Verify the functionality of the k8s config command during the smoke test.
    # It would be excessive to deploy a cluster solely for this purpose.    
    result = instance.exec(
        "k8s config --server 192.168.210.41".split(), capture_output=True
    )
    config = result.stdout.decode()
    assert len(config) > 0
    assert "server: https://192.168.210.41" in config
