#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
from pathlib import Path
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
        if len(channels) != 2:
            pytest.fail("'recent' requires the number of releases as second argument")
        _, num_channels = channels
        ref = config.GH_BASE_REF or config.GH_REF
        channels = snap.get_most_stable_channels(
            int(num_channels),
            config.FLAVOR,
            cp.arch,
            min_release=config.VERSION_UPGRADE_MIN_RELEASE,
            # Include `latest/edge/<flavor>` only if this is not a release branch.
            include_latest=ref == util.MAIN_BRANCH,
        )
        current_channel = channels[0]

    if config.SNAP:
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

    if len(channels) < 2:
        pytest.fail(
            f"Need at least 2 channels to upgrade, got {len(channels)} for flavor {config.FLAVOR}"
        )
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
            LOG.info(
                f"Upgrading {instance.id} from {current_channel} to channel {channel}"
            )

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
        if len(channels) != 2:
            pytest.fail("'recent' requires the number of releases as second argument")
        _, num_channels = channels
        ref = config.GH_BASE_REF or config.GH_REF
        channels = snap.get_most_stable_channels(
            int(num_channels),
            config.FLAVOR,
            cp.arch,
            min_release=config.VERSION_UPGRADE_MIN_RELEASE,
            reverse=True,
            # Include `latest/edge/<flavor>` only if this is not a release branch.
            include_latest=ref == util.MAIN_BRANCH,
        )
        if len(channels) < 2:
            pytest.fail(
                f"Need at least 2 channels to downgrade, got {len(channels)} for flavour {config.FLAVOR}"
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


@pytest.mark.node_count(4)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    # TODO(Adam): use TEST_VERSION_UPGRADE_CHANNELS if not set
    not config.SNAP,
    reason="Feature upgrades require a local snap file",
)
def test_feature_upgrades_inplace(instances: List[harness.Instance], tmp_path: Path):
    """Verify that feature upgrades function correctly.

    Note: This is an interim test that will be expanded as feature upgrades mature.
    Eventually, it will merge with test_version_upgrades to create a unified upgrade test.

    This test will spin up a three cp cluster on the previous track of the snap, and then upgrade to the snap.
    The test will then verify that the upgrade CR is updated correctly and that the features are upgraded
    after the last node is upgraded.
    The test will also verify that the feature version is not upgraded until all nodes are upgraded.
    """

    start_branch = util.previous_track(config.SNAP)
    bootstrap_cp = instances[0]
    worker = instances[-1]

    for instance in instances:
        instance.exec(f"snap install k8s --classic --channel={start_branch}".split())

    bootstrap_cp.exec(["k8s", "bootstrap"])
    for instance in instances:
        if instance.id in [bootstrap_cp.id, worker.id]:
            continue
        token = util.get_join_token(bootstrap_cp, instance)
        instance.exec(["k8s", "join-cluster", token])

    token = util.get_join_token(bootstrap_cp, worker, "--worker")
    worker.exec(["k8s", "join-cluster", token])

    # Get initial helm releases to track if they are updated correctly.
    initial_releases = {
        release["name"]: release
        for release in json.loads(
            bootstrap_cp.exec(
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

    # Refresh each CP node after each other and verify that the upgrade CR is updated correctly.
    for idx, instance in enumerate(instances):
        util.setup_k8s_snap(instance, tmp_path, config.SNAP)

        if instance.id == worker.id:
            continue

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

    # perform the final upgrade on the worker node.
    util.setup_k8s_snap(worker, tmp_path, config.SNAP)

    expected_instances = [instance.id for instance in instances]
    util.stubbornly(retries=15, delay_s=5).on(bootstrap_cp).until(
        lambda p: _waiting_for_upgraded_nodes(json.loads(p.stdout), expected_instances),
    ).exec(
        "k8s kubectl get upgrade -o=jsonpath={.items[0].status.upgradedNodes}".split(),
        capture_output=True,
        text=True,
    )

    # TODO(ben): Check that new fields are set in the feature config.
    # TODO(ben): Check that connectivity (e.g. for gateway) is working during the upgrade.

    util.stubbornly(retries=15, delay_s=5).on(bootstrap_cp).until(
        lambda p: p.stdout == "Completed",
    ).exec(
        "k8s kubectl get upgrade -o=jsonpath={.items[0].status.phase}".split(),
        capture_output=True,
        text=True,
    )

    p = bootstrap_cp.exec(
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

    current_releases = json.loads(p.stdout)
    for name, initial_rel in initial_releases.items():
        new_rel = None
        for r in current_releases:
            if r["name"] == name:
                new_rel = r
                break
        assert new_rel, f"Release {name} not in helm output"
        if initial_rel["updated"] == new_rel["updated"]:
            LOG.warning(
                "Release %s was not updated during upgrade. "
                "This might be due to a skipped helm apply due to same values or chart versions",
                name,
            )


def _waiting_for_upgraded_nodes(upgraded_nodes, expected_nodes) -> bool:
    LOG.info("Waiting for upgraded nodes %s to be: %s", upgraded_nodes, expected_nodes)
    return set(upgraded_nodes) == set(expected_nodes)


def _get_upgrade_crs(instance: harness.Instance) -> List[dict]:
    """Get the upgrade CRs of the cluster"""
    out = instance.exec(
        "k8s kubectl get upgrade -o=json".split(),
        capture_output=True,
        text=True,
    )
    return json.loads(out.stdout)["items"]


@pytest.mark.node_count(2)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.xfail(
    reason="The node removal does not work consistently due to a microcluster bug."
)
def test_feature_upgrades_rollout_upgrade(
    instances: List[harness.Instance], tmp_path: Path
):
    """ """
    # TODO: Ensure that this test only runs on different k8s versions.
    start_snap = util.previous_track(config.SNAP)
    main_old = instances[0]
    main_new = instances[3]

    # Setup the first half of nodes up on the old version.
    for instance in instances[:3]:
        instance.exec(f"snap install k8s --classic --channel={start_snap}".split())

    instance.exec(f"snap install k8s --classic --channel={start_snap}".split())

    main_old.exec(["k8s", "bootstrap"])
    for instance in instances[1:3]:
        token = util.get_join_token(main_old, instance)
        instance.exec(["k8s", "join-cluster", token])

    # Get initial helm releases to track if they are updated correctly.
    initial_releases = {
        release["name"]: release
        for release in json.loads(
            main_old.exec(
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

    # Add node with new version to the cluster
    # and remove an old one.
    for idx in range(3):
        new_instance = instances[3 + idx]
        cluster_node = instances[idx]

        util.setup_k8s_snap(new_instance, tmp_path, config.SNAP)
        token = util.get_join_token(cluster_node, new_instance)
        new_instance.exec(["k8s", "join-cluster", token])
        nodes_in_cluster = instances[idx : idx + 3]  # noqa
        util.wait_until_k8s_ready(new_instance, nodes_in_cluster)

        # An upgrade CRD should exist and be in NodeUpgrade phase.
        crs = _get_upgrade_crs(new_instance)
        assert len(crs) == 1, f"Expected one upgrade CR but got {crs}"
        assert (
            crs[0]["status"]["phase"] == "NodeUpgrade"
        ), f"Expected NodeUpgrade but got {crs[0]['status']['phase']}"

        # Remove old node from cluster
        new_instance.exec(["k8s", "remove-node", cluster_node.id])

    # After all nodes are upgraded, the phase should be FeatureUpgrade/Completed
    # and the helm releases should be updated.
    util.stubbornly(retries=15, delay_s=5).on(main_new).until(
        lambda p: p.stdout == "Completed",
    ).exec(
        "k8s kubectl get upgrade -o=jsonpath={.items[0].status.phase}".split(),
        capture_output=True,
        text=True,
    )

    # # All Feature version should eventually be upgraded.
    LOG.info("Waiting for all helm releases to upgrade")
    util.stubbornly(retries=15, delay_s=5).on(main_new).until(
        lambda p: all(
            next(r for r in json.loads(p.stdout) if r["name"] == name)["updated"]
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
