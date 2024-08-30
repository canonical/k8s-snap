#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import re
import subprocess
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
def test_upgrades(instances: List[harness.Instance]):
    instance = instances[0]

    # Channels to which the cluster will be upgraded.
    # First entry is the bootstrap channel.
    # Afterwards, upgrades are done in order.
    # TODO: Move this to env so that we can configure that for moonray.
    channels = [
        "1.30/candidate",
        "latest/edge",
    ]

    # Log the current snap revisions for the k8s snap.
    instance.exec(["snap", "info", "k8s"])

    # Setup the k8s snap from the bootstrap channel and setup basic configuration.
    # TODO: make this a cluster
    instance.exec(["snap", "install", "k8s", "--channel", channels[0]])
    instance.exec(["k8s", "bootstrap"])

    # Create a join token for workers (https://github.com/canonical/k8s-snap/issues/634)
    instance.exec(["k8s", "get-join-token", "test", "--worker"])
    instance.exec(["k8s", "status", "--wait-ready"])

    current_channel = channels[0]
    for channel in channels[1:]:
        LOG.info(f"Upgrading from {current_channel} to channel {channel}")
        instance.exec(["snap", "refresh", "k8s", "--channel", channel])

        # Verify that the upgrade was successful.
        instance.exec(["k8s", "status", "--wait-ready"])
