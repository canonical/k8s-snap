#
# Copyright 2026 Canonical, Ltd.
#
from unittest.mock import patch

import pytest
from test_util.harness import HarnessError
from test_util.harness.juju import JujuHarness


def test_juju_harness_allows_ipv4_and_rejects_other_network_types():
    with (
        patch("test_util.harness.juju.config.JUJU_CONTROLLER", "ctrl"),
        patch("test_util.harness.juju.config.JUJU_MODEL", "mdl"),
        patch("test_util.harness.juju.config.JUJU_MACHINES", "0"),
    ):
        harness = JujuHarness()

    instance = harness.new_instance()
    assert instance.id == "0"

    with pytest.raises(HarnessError, match="only IPv4 is supported"):
        harness.new_instance(network_type="dualstack")
