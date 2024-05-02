#
# Copyright 2024 Canonical, Ltd.
#
import logging

from test_util import harness

LOG = logging.getLogger(__name__)


def test_smoke(instance: harness.Instance):
    # Verify the functionality of the k8s config command during the smoke test.
    # It would be excessive to deploy a cluster solely for this purpose.
    result = instance.exec(
        "k8s config --server 192.168.210.41".split(), capture_output=True
    )
    config = result.stdout.decode()
    assert len(config) > 0
    assert "server: https://192.168.210.41" in config
