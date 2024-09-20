#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
#@pytest.mark.xfail("cilium failures are blocking this from working")
@pytest.mark.skipif(
    not config.VERSION_UPGRADE_CHANNELS, reason="No upgrade channels configured"
)
def test_version_upgrades(instances: List[harness.Instance]):
    channels = config.VERSION_UPGRADE_CHANNELS
    cp = instances[0]

    if channels[0].lower() == "recent":
        if len(channel) != 3:
            pytest.fail("'recent' requires the number of releases as second argument and the flavour as third argument")
        num_channels = int(channels[1])
        flavour = channels[2]
        channels = util.get_latest_channels(num_channels, flavour)


    LOG.info(f"Bootstrap node on {channels[0]} and upgrade through channels: {channels[1:]}")

    # Setup the k8s snap from the bootstrap channel and setup basic configuration.
    cp.exec(["snap", "install", "k8s", "--channel", channels[0]])
    cp.exec(["k8s", "bootstrap"])

    util.stubbornly(retries=30, delay_s=20).until(util.ready_nodes(cp) == 1)

    current_channel = channels[0]
    for channel in channels[1:]:
        LOG.info(
            f"Upgrading {cp.id} from {current_channel} to channel {channel}"
        )
        # Log the current snap version on the node.
        cp.exec(["snap", "info", "k8s"])

        # note: the `--classic` flag will be ignored by snapd for strict snaps.
        cp.exec(
            ["snap", "refresh", "k8s", "--channel", channel, "--classic", "--amend"]
        )

        util.stubbornly(retries=30, delay_s=20).until(util.ready_nodes(cp) == 3)
        LOG.info(f"Upgraded {cp.id} to channel {channel}")