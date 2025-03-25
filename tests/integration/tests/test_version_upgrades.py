#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import os
from pathlib import Path
import time
from typing import List

import pytest
import yaml
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

    # Copy the current snap into the instances.
    snap_path = (tmp_path / "k8s.snap").as_posix()
    for instance in instances:
        instance.send_file(config.SNAP, snap_path)

    # Figure out where to add the current snap into the channels array.
    # Upgrades should be in order.
    out = cp.exec(["snap", "info", snap_path], capture_output=True)
    info = yaml.safe_load(out.stdout)

    # expected: "v1.32.2 classic"
    ver = info["version"].lstrip("v").split()[0].split(".")
    added = False
    for i in range(len(channels)):
        # e.g.: 1.32-classic/stable
        chan_ver = channels[i].split("-")[0].split(".")
        if len(chan_ver) > 1 and (ver[0], ver[1]) < (chan_ver[0], chan_ver[1]):
            channels.insert(i, snap_path)
            added = True
            break

    if not added:
        # if not added yet, config.SNAP should be at the end.
        channels.append(snap_path)

    LOG.info(f"Testing upgrades for snaps: {channels}")
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

    LOG.info(f"Installed {len(instances)} nodes on channel {current_channel}")

    for channel in channels[1:]:
        for instance in instances:
            LOG.info(f"Upgrading {instance.id} from {current_channel} to {channel}")

            # Log the current snap version on the node.
            out = instance.exec(["snap", "list", config.SNAP_NAME], capture_output=True)
            latest_version = out.stdout.decode().strip().split("\n")[-1]
            LOG.info(f"Current snap version: {latest_version}")

            # note: the `--classic` flag will be ignored by snapd for strict snaps.
            cmd = [
                "snap",
                "refresh",
                config.SNAP_NAME,
                "--channel",
                channel,
                "--classic",
            ]
            if channel.startswith("/"):
                LOG.info("Refreshing k8s snap by path")
                cmd = ["snap", "install", "--classic", "--dangerous", snap_path]

            instance.exec(cmd)
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


@pytest.mark.node_count(3)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
def test_feature_upgrades(instances: List[harness.Instance], tmp_path: Path):
    """Test the feature upgrades work
    Note(ben): This test is a work in progress and will
    be expanded with new tests as the feature upgrades work takes shape.
    Once this work is complete, this test will likely be merged with the
    test_version_upgrades test above and create a single test for all upgrades.

    It is not possible to:
        * do a snap refresh to a local snap
        * upload the same snap revision to different branches
    Therefore we need to:
        * upload the snap to two different branches and perform an upgrade between them
          (we rely on/need to test pre-refresh hooks so this cannot be done by upgrading from an existing revision)
        * Need to do a dummy modification to the snap to make it possible to upload the same revision to different branches

    """
    assert (
        config.SNAPCRAFT_STORE_CREDENTIALS is not None
    ), "SNAPCRAFT_STORE_CREDENTIALS must be set to run this test"

    assert config.SNAP is not None, "SNAP must be set to run this test"

    start_branch = "latest/edge/ci-upgrade-test-1"
    target_branch = "latest/edge/ci-upgrade-test-2"

    os.environ["SNAPCRAFT_STORE_CREDENTIALS"] = config.SNAPCRAFT_STORE_CREDENTIALS

    # unsquash, add dummy change to ensure uniqueness in store, squash, upload to branches
    # note: we also add a dummy change to the first branch as the test would otherwise fail if a PR only introduces test changes.
    unsquash_path = tmp_path / "k8s-snap-unsquashed"
    util.run(f"unsquashfs -d {unsquash_path} {config.SNAP}".split())
    for idx, branch in enumerate([start_branch, target_branch]):
        # create a random dummy file to ensure the snap is unique
        util.run(f"touch { unsquash_path/f"{time.time()}" }".split())
        util.run(
            f"snapcraft pack {unsquash_path} -o {tmp_path / f"k8s-snap-modified-{idx+1}.snap"}".split()
        )
        util.run(
            f"snapcraft upload { tmp_path / f"k8s-snap-modified-{idx+1}.snap"} --release={branch}".split()
        )

    main = instances[0]

    for instance in instances:
        instance.exec(f"snap install k8s --classic --channel={start_branch}".split())

    main.exec(["k8s", "bootstrap"])
    for instance in instances[1:]:
        token = util.get_join_token(main, instance)
        instance.exec(["k8s", "join-cluster", token])

    util.wait_until_k8s_ready(instance, instances)

    # Verify that the UpgradeCRD is known to the cluster
    main.exec("k8s kubectl get crd upgrades.k8sd.io".split())

    # Refresh each node after eachother and verify that the upgrade CR is updated correctly.
    for idx, instance in enumerate(instances):
        instance.exec(f"snap refresh k8s --channel={target_branch}".split())

        # TODO(ben): Check if this wait is really required, if yes - why?
        expected_instances = [instance.id for instance in instances[: idx + 1]]
        util.stubbornly(retries=15, delay_s=5).on(instance).until(
            lambda p: _waiting_for_upgraded_nodes(
                json.loads(p.stdout), expected_instances
            ),
        ).exec(
            "k8s kubectl get upgrade -o=jsonpath={.items[0].status.upgradedNodes}".split(),
            capture_output=True,
            text=True,
        )

        phase = instance.exec(
            "k8s kubectl get upgrade -o=jsonpath={.items[0].status.phase}".split(),
            capture_output=True,
            text=True,
        ).stdout

        if idx == len(instances) - 1:
            assert (
                phase == "FeatureUpgrade"
            ), f"After the last upgrade, expected phase to be FeatureUpgrade but got {phase}"
        else:
            assert (
                phase == "NodeUpgrade"
            ), f"While upgrading, expected phase to be NodeUpgrade but got {phase}"


def _waiting_for_upgraded_nodes(upgraded_nodes, expected_nodes) -> True:
    LOG.info("Waiting for upgraded nodes %s to be: %s", upgraded_nodes, expected_nodes)
    return set(upgraded_nodes) == set(expected_nodes)
