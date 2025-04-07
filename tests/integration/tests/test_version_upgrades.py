#
# Copyright 2025 Canonical, Ltd.
#
import logging
from pathlib import Path
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
        LOG.info(f"Current snap version: {out.stdout.decode().strip()}")

        # note: the `--classic` flag will be ignored by snapd for strict snaps.
        cp.exec(
            ["snap", "refresh", config.SNAP_NAME, "--channel", channel, "--classic"]
        )
        util.wait_until_k8s_ready(cp, instances)
        current_channel = channel
        LOG.info(f"Upgraded {cp.id} on channel {channel}")


@pytest.mark.tags(tags.NIGHTLY)
def test_feature_upgrades(instances: List[harness.Instance]):
    """Test the feature upgrades work
    Note(ben): This test is a work in progress and will
    be expanded with new tests as the feature upgrades work takes shape.
    Once this work is complete, this test will likely be merged with the
    test_version_upgrades test above and create a single test for all upgrades."""

    cp = instances[0]

    # Verify that the UpgradeCRD is known to the cluster
    cp.exec("k8s kubectl get crd upgrades.k8sd.io".split())

    # Test that the UpgradeCRD can be created.
    cp.exec(
        "k8s kubectl apply -f -".split(),
        input=str.encode(Path(config.MANIFESTS_DIR / "upgrade.yaml").read_text()),
    )
    cp.exec("k8s kubectl get upgrade cluster-upgrade".split())
