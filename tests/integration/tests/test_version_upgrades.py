#
# Copyright 2025 Canonical, Ltd.
#
import json
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


@pytest.mark.node_count(3)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
def test_feature_upgrades(instances: List[harness.Instance], tmp_path: Path):
    """Verify that feature upgrades function correctly.

    Note: This is an interim test that will be expanded as feature upgrades mature.
    Eventually, it will merge with test_version_upgrades to create a unified upgrade test.

    This test will spin up a three cp cluster on 1.32-classic/stable, and then upgrade to the snap.
    The test will then verify that the upgrade CR is updated correctly and that the features are upgraded
    after the last node is upgraded.
    The test will also verify that the feature version is not upgraded until all nodes are upgraded.
    """
    assert config.SNAP is not None, "SNAP must be set to run this test"

    start_branch = util.previous_track(config.SNAP)
    main = instances[0]

    for instance in instances:
        instance.exec(f"snap install k8s --classic --channel={start_branch}".split())

    main.exec(["k8s", "bootstrap"])
    for instance in instances[1:]:
        token = util.get_join_token(main, instance)
        instance.exec(["k8s", "join-cluster", token])

    # Get initial helm releases to track if they are updated correctly.
    initial_releases = {
        release["name"]: release
        for release in json.loads(
            main.exec(
                [
                    "/snap/k8s/current/bin/helm",
                    "--kubeconfig",
                    "/etc/kubernetes/admin.conf",
                    "list",
                    "-n",
                    "kube-system",
                    "-o",
                    "json",
                ],
                capture_output=True,
                text=True,
            ).stdout
        )
    }

    # Refresh each node after each other and verify that the upgrade CR is updated correctly.
    for idx, instance in enumerate(instances):
        util.setup_k8s_snap(instance, tmp_path, config.SNAP)

        # The crd will be created once the node is up and ready, so we might need to wait for it.
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
            util.stubbornly(retries=15, delay_s=5).on(instance).until(
                lambda p: p.stdout in ["FeatureUpgrade", "Completed"],
            ).exec(
                "k8s kubectl get upgrade -o=jsonpath={.items[0].status.phase}".split(),
                capture_output=True,
                text=True,
            )

            # All Feature version should eventually be upgraded.
            LOG.info("Waiting for all helm releases to upgrade")
            util.stubbornly(retries=15, delay_s=5).on(instance).until(
                lambda p: all(
                    next(r for r in json.loads(p.stdout) if r["name"] == name)[
                        "updated"
                    ]
                    != initial_releases[name]["updated"]
                    for name in initial_releases
                ),
            ).exec(
                [
                    "/snap/k8s/current/bin/helm",
                    "--kubeconfig",
                    "/etc/kubernetes/admin.conf",
                    "list",
                    "-n",
                    "kube-system",
                    "-o",
                    "json",
                ],
                capture_output=True,
                text=True,
            )
            LOG.info("All helm releases have upgraded successfully")

            # TODO(ben): Check that new fields are set in the feature config.
            # TODO(ben): Check that connectivity (e.g. for gateway) is working during the upgrade.

            util.stubbornly(retries=15, delay_s=5).on(instance).until(
                lambda p: p.stdout == "Completed",
            ).exec(
                "k8s kubectl get upgrade -o=jsonpath={.items[0].status.phase}".split(),
                capture_output=True,
                text=True,
            )
        else:
            assert (
                phase == "NodeUpgrade"
            ), f"While upgrading, expected phase to be NodeUpgrade but got {phase}"

            current_helm_releases = instance.exec(
                [
                    "/snap/k8s/current/bin/helm",
                    "--kubeconfig",
                    "/etc/kubernetes/admin.conf",
                    "list",
                    "-n",
                    "kube-system",
                    "-o",
                    "json",
                ],
                capture_output=True,
                text=True,
            ).stdout

            for release in json.loads(current_helm_releases):
                LOG.info(json.dumps(json.loads(current_helm_releases), indent=2))
                LOG.info("Checking helm release %s", release["name"])
                name = release["name"]
                assert (
                    release["updated"] == initial_releases[name]["updated"]
                ), f"{release['name']} was updated while upgrading {instance.id} but should not \
                    have been ({initial_releases[name]['updated']}, {release['updated']})"


def _waiting_for_upgraded_nodes(upgraded_nodes, expected_nodes) -> True:
    LOG.info("Waiting for upgraded nodes %s to be: %s", upgraded_nodes, expected_nodes)
    return set(upgraded_nodes) == set(expected_nodes)
