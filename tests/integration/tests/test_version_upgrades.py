#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, snap, tags, util
from test_util.registry import Registry

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(4)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_UPGRADE_CHANNELS, reason="No upgrade channels configured"
)
@pytest.mark.tags(tags.NIGHTLY)
def test_version_upgrades(
    instances: List[harness.Instance],
    tmp_path,
    containerd_cfgdir: str,
    registry: Registry,
):
    channels = config.VERSION_UPGRADE_CHANNELS
    cp = instances[0]
    cp1 = instances[1]
    cp2 = instances[2]
    w0 = instances[3]
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
    for instance in instances:
        util.setup_k8s_snap(instance, tmp_path, current_channel)
        if config.USE_LOCAL_MIRROR:
            registry.apply_configuration(instance, containerd_cfgdir)

    cp.exec(["k8s", "bootstrap"])

    join_token_cp1 = util.get_join_token(cp, cp1)
    join_token_cp2 = util.get_join_token(cp, cp2)
    join_token_w0 = util.get_join_token(cp, w0, "--worker")

    util.join_cluster(cp1, join_token_cp1)
    util.join_cluster(cp2, join_token_cp2)
    util.join_cluster(w0, join_token_w0)

    util.wait_until_k8s_ready(cp, instances)
    nodes = util.ready_nodes(cp)
    assert len(nodes) == 4, "all nodes should have joined cluster"

    LOG.info(f"Installed {len(instances)} nodes on channel {current_channel}")

    for channel in channels[1:]:
        for instance in instances:
            LOG.info(
                f"Upgrading {instance.id} from {current_channel} to channel {channel}"
            )

            # Log the current snap version on the node.
            out = instance.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
            latest_version = out.stdout.decode().strip().split("\n")[-1]
            LOG.info(f"Current snap version: {latest_version}")

            # note: the `--classic` flag will be ignored by snapd for strict snaps.
            instance.exec(
                ["snap", "refresh", config.SNAP_NAME, "--channel", channel, "--classic"]
            )
            util.wait_until_k8s_ready(cp, instances)
            current_channel = channel
            LOG.info(f"Upgraded {instance.id} on channel {channel}")


@pytest.mark.node_count(3)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_DOWNGRADE_CHANNELS, reason="No downgrade channels configured"
)
@pytest.mark.tags(tags.NIGHTLY)
def test_version_downgrades_with_rollback(
    instances: List[harness.Instance],
    tmp_path,
    containerd_cfgdir: str,
    registry: Registry,
):
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
    cp1 = instances[1]
    cp2 = instances[2]
    # TODO: add a worker node once the snap refresh is fixed on worker nodes
    # and the patch lands on all the release channels covered by this test.
    #
    # At the moment, the following fails:
    # https://github.com/canonical/k8s-snap/blob/96124bd7f1e82e96e23a4c4d11fcff86045f403c/snap/hooks/configure#L7
    #
    # w0 = instances[3]
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
    for instance in instances:
        util.setup_k8s_snap(instance, tmp_path, current_channel)
        if config.USE_LOCAL_MIRROR:
            registry.apply_configuration(instance, containerd_cfgdir)

    cp.exec(["k8s", "bootstrap"])

    join_token_cp1 = util.get_join_token(cp, cp1)
    join_token_cp2 = util.get_join_token(cp, cp2)
    # join_token_w0 = util.get_join_token(cp, w0, "--worker")

    util.join_cluster(cp1, join_token_cp1)
    util.join_cluster(cp2, join_token_cp2)
    # util.join_cluster(w0, join_token_w0)

    util.wait_until_k8s_ready(cp, instances)
    nodes = util.ready_nodes(cp)
    assert len(nodes) == len(instances), "all nodes should have joined cluster"

    for channel in channels[1:]:
        for instance in instances:
            LOG.info(
                "Initiating downgrade + rollback segment from "
                f"{current_channel} → {channel} - {instance.id}"
            )
            out = instance.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
            latest_version = out.stdout.decode().strip().split("\n")[-1]
            LOG.info(f"Current snap version: {latest_version}")

            LOG.debug(
                f"Step 1. Downgrade {instance.id} from {current_channel} → {channel}"
            )
            # note: the `--classic` flag will be ignored by snapd for strict snaps.
            instance.exec(
                ["snap", "refresh", config.SNAP_NAME, "--channel", channel, "--classic"]
            )
            util.wait_until_k8s_ready(cp, instances)

        last_channel = current_channel
        current_channel = channel

        for instance in instances:
            LOG.debug(f"Step 2. Roll back from {current_channel} → {last_channel}")
            # note: the `--classic` flag will be ignored by snapd for strict snaps.
            instance.exec(
                [
                    "snap",
                    "refresh",
                    config.SNAP_NAME,
                    "--channel",
                    last_channel,
                    "--classic",
                ]
            )
            util.wait_until_k8s_ready(cp, instances)

        for instance in instances:
            LOG.debug(
                f"Step 3. Final downgrade to channel from {last_channel} → {current_channel}"
            )
            instance.exec(
                [
                    "snap",
                    "refresh",
                    config.SNAP_NAME,
                    "--channel",
                    current_channel,
                    "--classic",
                ]
            )
            util.wait_until_k8s_ready(cp, instances)

            LOG.info("Rollback segment complete. Proceeding to next downgrade segment.")

    LOG.info("Rollback test complete. All downgrade segments verified.")
