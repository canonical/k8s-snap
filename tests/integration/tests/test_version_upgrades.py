#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, snap, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_UPGRADE_CHANNELS, reason="No upgrade channels configured"
)
@pytest.mark.tags(tags.NIGHTLY)
def test_version_upgrades(instances: List[harness.Instance], tmp_path):
    channels = config.VERSION_UPGRADE_CHANNELS
    cp = instances[0]
    current_channel = channels[0]

    if current_channel.lower() == "recent":
        if len(channels) != 3:
            pytest.fail(
                "'recent' requires the number of releases as second argument and the flavour as third argument"
            )
        _, num_channels, flavour = channels
        # Currently, it fails to refresh the snap to the 1.33 channel or newer.
        # Just test upgrade to at most 1.32.
        channels = snap.get_most_stable_channels(
            int(num_channels),
            flavour,
            cp.arch,
            include_latest=False,
            min_release=config.VERSION_UPGRADE_MIN_RELEASE,
            max_release="1.32",
        )
        if len(channels) < 2:
            pytest.fail(
                f"Need at least 2 channels to upgrade, got {len(channels)} for flavour {flavour}"
            )
        current_channel = channels[0]

    LOG.info(
        f"Bootstrap node on {current_channel} and upgrade through channels: {channels[1:]}"
    )

    # Setup the k8s snap from the bootstrap channel and setup basic configuration.
    util.setup_k8s_snap(cp, tmp_path, current_channel)
    cp.exec(["k8s", "bootstrap"])

    util.wait_until_k8s_ready(cp, instances)
    LOG.info(f"Installed {cp.id} on channel {current_channel}")

    for channel in channels[1:]:
        LOG.info(f"Upgrading {cp.id} from {current_channel} to channel {channel}")

        # Log the current snap version on the node.
        out = cp.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
        latest_version = out.stdout.decode().strip().split("\n")[-1]
        LOG.info(f"Current snap version: {latest_version}")

        # note: the `--classic` flag will be ignored by snapd for strict snaps.
        cp.exec(
            ["snap", "refresh", config.SNAP_NAME, "--channel", channel, "--classic"]
        )
        util.wait_until_k8s_ready(cp, instances)
        current_channel = channel
        LOG.info(f"Upgraded {cp.id} on channel {channel}")


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_DOWNGRADE_CHANNELS, reason="No downgrade channels configured"
)
@pytest.mark.tags(tags.NIGHTLY)
def test_version_downgrades_with_rollback(instances: List[harness.Instance], tmp_path):
    """
    This test will downgrade the snap through the channels, and at each downgrade, attempt a rollback.

    Example of downgrading while rolling back through channels:
    Channels from config:  1.32-classic/stable, 1.31-classic/stable
    Segment 1: 1.32-classic/stable -> 1.31-classic/stable -> 1.32-classic/stable -> 1.31-classic/stable

    Example 2 of downgrading while rolling back through channels:
    Channels from config: 1.32-classic/stable 1.32-classic/beta 1.31-classic/stable
    Segment 1: 1.32-classic/stable -> 1.32-classic/beta -> 1.32-classic/stable -> 1.32-classic/beta
    Segment 2: 1.32-classic/beta -> 1.31-classic/stable -> 1.32-classic/beta -> 1.31-classic/stable
    """
    channels = config.VERSION_DOWNGRADE_CHANNELS
    cp = instances[0]
    current_channel = channels[0]

    if current_channel.lower() == "recent":
        if len(channels) != 3:
            pytest.fail(
                "'recent' requires the number of releases as second argument and the flavour as third argument"
            )
        _, num_channels, flavour = channels
        channels = snap.get_most_stable_channels(
            int(num_channels),
            flavour,
            cp.arch,
            min_release=config.VERSION_UPGRADE_MIN_RELEASE,
            reverse=True,
            include_latest=False,
        )
        if len(channels) < 2:
            pytest.fail(
                f"Need at least 2 channels to downgrade, got {len(channels)} for flavour {flavour}"
            )
        current_channel = channels[0]

    LOG.info(
        f"Bootstrap node on {current_channel} and downgrade through channels: {channels[1:]}"
    )

    # Setup the k8s snap from the bootstrap channel and setup basic configuration.
    util.setup_k8s_snap(cp, tmp_path, current_channel)
    cp.exec(["k8s", "bootstrap"])

    util.wait_until_k8s_ready(cp, instances)
    LOG.info(f"Installed {cp.id} on channel {current_channel}")

    for channel in channels[1:]:
        LOG.info(
            f">>> Initiating downgrade + rollback segment from {current_channel} → {channel}"
        )
        out = cp.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
        latest_version = out.stdout.decode().strip().split("\n")[-1]
        LOG.info(f"Current snap version: {latest_version}")

        LOG.info(f"Step 1. Downgrade {cp.id} from {current_channel} → {channel}")
        # note: the `--classic` flag will be ignored by snapd for strict snaps.
        cp.exec(
            ["snap", "refresh", config.SNAP_NAME, "--channel", channel, "--classic"]
        )
        util.wait_until_k8s_ready(cp, instances, output=False)

        last_channel = current_channel
        current_channel = channel

        LOG.info(f"Step 2. Roll back from {current_channel} → {last_channel}")
        # note: the `--classic` flag will be ignored by snapd for strict snaps.
        cp.exec(
            [
                "snap",
                "refresh",
                config.SNAP_NAME,
                "--channel",
                last_channel,
                "--classic",
            ]
        )
        util.wait_until_k8s_ready(cp, instances, output=False)

        LOG.info(
            f"Step 3. Final downgrade to channel from {last_channel} → {current_channel}"
        )
        cp.exec(
            [
                "snap",
                "refresh",
                config.SNAP_NAME,
                "--channel",
                current_channel,
                "--classic",
            ]
        )
        util.wait_until_k8s_ready(cp, instances, output=False)

        if channel == channels[-1]:
            LOG.info(">>> Rollback test complete. All downgrade segments verified.")
        else:
            LOG.info(
                ">>> Rollback segment complete. Proceeding to next downgrade segment."
            )
