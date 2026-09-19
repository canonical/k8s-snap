#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import subprocess
import time
from pathlib import Path
from typing import List

import pytest
import yaml
from tenacity import stop_after_delay
from test_util import config, harness, snap, tags, util
from test_util.registry import Registry

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(4)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_UPGRADE_CHANNELS, reason="No upgrade channels configured"
)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="runner size too small on multipass"
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
        LOG.info(f"Locally built snap version: {ver}")
        local_snap_version = (int(ver[0]), int(ver[1]))
        added = False

        for i in range(len(channels)):
            if "latest" in channels[i]:
                continue

            # e.g.: 1.32-classic/stable
            chan_ver_parts = channels[i].split("-")[0].split(".")
            chan_ver = (int(chan_ver_parts[0]), int(chan_ver_parts[1]))
            if local_snap_version < chan_ver:
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
        util.setup_k8s_snap(instance, current_channel)
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

            if channel.startswith("/"):
                LOG.info("Refreshing k8s snap by path")
                instance.exec(
                    ["snap", "install", "--classic", "--dangerous", snap_path]
                )
            else:
                util.snap_refresh(instance, channel, "--amend")
            util.wait_until_k8s_ready(cp, instances)
            LOG.info("Verifying snap service health")
            util.check_snap_services_ready(instance, retries=10, delay_s=10)
            util.check_service_restarts(instance)
            util.check_service_logs_for_panics(instance)
            LOG.info(f"Upgraded {instance.id} to channel {channel}")

        current_channel = channel
        LOG.info(f"Upgraded all instances to channel {channel}")

        LOG.info("Waiting for all pods to be ready after upgrade")
        util.wait_for_pods_ready(cp)
        LOG.info("All pods are ready after upgrade")


@pytest.mark.node_count(3)
@pytest.mark.no_setup()
@pytest.mark.skipif(
    not config.VERSION_DOWNGRADE_CHANNELS, reason="No downgrade channels configured"
)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="runner size too small on multipass"
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
        max_release = (
            ref.removeprefix("release-") if ref and ref.startswith("release-") else None
        )
        channels = snap.get_most_stable_channels(
            int(num_channels),
            config.FLAVOR,
            cp.arch,
            min_release=config.VERSION_UPGRADE_MIN_RELEASE,
            max_release=max_release,
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
        util.setup_k8s_snap(instance, current_channel)
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
            util.snap_refresh(instance, channel)
            util.wait_until_k8s_ready(cp, instances)
            LOG.info("Verifying snap service health")
            util.check_snap_services_ready(instance, retries=10, delay_s=10)
            util.check_service_restarts(instance)
            util.check_service_logs_for_panics(instance)

        last_channel = current_channel
        current_channel = channel

        for instance in instances:
            LOG.debug(f"Step 2. Roll back from {current_channel} → {last_channel}")
            util.snap_refresh(instance, last_channel)
            util.wait_until_k8s_ready(cp, instances)
            LOG.info("Verifying snap service health")
            util.check_snap_services_ready(instance, retries=10, delay_s=10)
            util.check_service_restarts(instance)
            util.check_service_logs_for_panics(instance)

        for instance in instances:
            LOG.debug(
                f"Step 3. Final downgrade to channel from {last_channel} → {current_channel}"
            )
            util.snap_refresh(instance, current_channel)
            util.wait_until_k8s_ready(cp, instances)
            LOG.info("Verifying snap service health")
            util.check_snap_services_ready(instance, retries=10, delay_s=10)
            util.check_service_restarts(instance)
            util.check_service_logs_for_panics(instance)

            LOG.info("Rollback segment complete. Proceeding to next downgrade segment.")

        LOG.info("Waiting for all pods to be ready after downgrade segment")
        # Use a generous timeout for pod readiness after downgrade segments.
        # Multiple rapid version transitions (downgrade + rollback + final downgrade
        # across all nodes) can leave pods in a degraded state that takes longer than
        # the default 10 minutes to recover, especially on arm64.
        util.wait_for_pods_ready(cp, retries=180, delay_s=5)
        LOG.info("All pods are ready after downgrade segment")

    LOG.info("Rollback test complete. All downgrade segments verified.")


@pytest.mark.node_count(4)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="runner size too small on multipass"
)
@pytest.mark.skipif(
    # TODO(Adam): use TEST_VERSION_UPGRADE_CHANNELS if not set
    not config.SNAP,
    reason="Feature upgrades require a local snap file",
)
def test_feature_upgrades_inplace(
    instances: List[harness.Instance], tmp_path: Path, request
):
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
        util.stubbornly(retries=3, delay_s=30).on(instance).exec(
            [
                "snap",
                "install",
                "k8s",
                "--classic",
                *util.snap_channel_args(start_branch),
            ]
        )

    bootstrap_cp.exec(["k8s", "bootstrap"])
    for instance in instances:
        if instance.id in [bootstrap_cp.id, worker.id]:
            continue
        token = util.get_join_token(bootstrap_cp, instance)
        instance.exec(["k8s", "join-cluster", token])

    token = util.get_join_token(bootstrap_cp, worker, "--worker")
    worker.exec(["k8s", "join-cluster", token])

    initial_dns = _coredns_deployment(bootstrap_cp)
    LOG.info("CoreDNS policy before upgrade: %s", initial_dns["spec"])
    _wait_coredns_replicas(bootstrap_cp, initial_dns["spec"].get("replicas", 1))
    _start_dns_upgrade_probe(bootstrap_cp, request)

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
        if instance.id == worker.id:
            continue

        util.setup_k8s_snap(instance, config.SNAP)
        _assert_dns_upgrade_probe(bootstrap_cp)

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
    util.setup_k8s_snap(worker, config.SNAP)

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

    try:
        util.stubbornly(retries=15, delay_s=5, reraise=True).on(bootstrap_cp).until(
            lambda result: _upgrade_completed(json.loads(result.stdout)),
        ).exec(
            ["k8s", "kubectl", "get", "upgrade", "-o=json", "--request-timeout=10s"],
            text=True,
            timeout=20,
        )
    except (AssertionError, subprocess.SubprocessError):
        for instance in instances:
            try:
                result = instance.exec(
                    [
                        "journalctl",
                        "-u",
                        "snap.k8s.k8sd",
                        "--since",
                        "10 minutes ago",
                        "--no-pager",
                        "-n",
                        "200",
                    ],
                    capture_output=True,
                    text=True,
                    check=False,
                    timeout=20,
                )
                LOG.error(
                    "Upgrade failure on %s:\n%s\n%s",
                    instance.id,
                    result.stdout,
                    result.stderr,
                )
            except (OSError, subprocess.SubprocessError) as diagnostic_error:
                LOG.warning(
                    "Could not collect upgrade diagnostics from %s: %s",
                    instance.id,
                    diagnostic_error,
                )
        raise

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

    LOG.info("Waiting for all pods to be ready after upgrade")
    util.wait_for_pods_ready(bootstrap_cp)
    LOG.info("All pods are ready after upgrade")

    upgraded_dns = _coredns_deployment(bootstrap_cp)
    _assert_coredns_scheduling_policy(upgraded_dns)
    _assert_dns_upgrade_probe(bootstrap_cp)
    LOG.info(
        "CoreDNS deployment generation before/after upgrade: %s -> %s",
        initial_dns["metadata"]["generation"],
        upgraded_dns["metadata"]["generation"],
    )
    _check_coredns_hpa_scaling(bootstrap_cp, instances)
    _assert_dns_upgrade_probe(bootstrap_cp)


def _start_dns_upgrade_probe(instance: harness.Instance, request):
    probe = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "dns-upgrade-probe", "namespace": "default"},
        "spec": {
            "restartPolicy": "Never",
            "activeDeadlineSeconds": 2700,
            "containers": [
                {
                    "name": "probe",
                    "image": "ghcr.io/containerd/busybox:1.28",
                    "command": [
                        "sh",
                        "-c",
                        "until timeout -t 3 nslookup kubernetes.default.svc.cluster.local "
                        ">/tmp/dns-lookup.log 2>&1; do sleep 2; done; "
                        "touch /tmp/ready; "
                        "while true; do "
                        "if timeout -t 3 nslookup kubernetes.default.svc.cluster.local >/tmp/dns-lookup.log 2>&1; "
                        "then echo OK $(date +%s); else echo FAIL $(date +%s); exit 1; fi; sleep 2; done",
                    ],
                    "readinessProbe": {
                        "exec": {"command": ["test", "-f", "/tmp/ready"]},
                        "periodSeconds": 2,
                    },
                }
            ],
        },
    }
    request.addfinalizer(
        lambda: instance.exec(
            [
                "k8s",
                "kubectl",
                "delete",
                "pod",
                "dns-upgrade-probe",
                "--ignore-not-found",
                "--wait=false",
                "--request-timeout=10s",
            ],
            check=False,
            timeout=20,
        )
    )
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-", "--request-timeout=10s"],
        input=json.dumps(probe),
        text=True,
        timeout=20,
    )
    try:
        instance.exec(
            [
                "k8s",
                "kubectl",
                "wait",
                "pod/dns-upgrade-probe",
                "--for=condition=Ready",
                "--timeout=120s",
            ],
            timeout=130,
        )
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        for arguments in (
            ["describe", "pod", "dns-upgrade-probe"],
            ["logs", "dns-upgrade-probe", "--tail=50"],
            ["exec", "dns-upgrade-probe", "--", "cat", "/tmp/dns-lookup.log"],
        ):
            try:
                result = instance.exec(
                    ["k8s", "kubectl", "--request-timeout=10s", *arguments],
                    capture_output=True,
                    text=True,
                    check=False,
                    timeout=20,
                )
                LOG.error(
                    "DNS probe diagnostics (%s):\n%s\n%s",
                    arguments,
                    result.stdout,
                    result.stderr,
                )
            except (OSError, subprocess.SubprocessError) as diagnostic_error:
                LOG.warning(
                    "Could not collect DNS probe diagnostics: %s", diagnostic_error
                )
        raise
    _assert_dns_upgrade_probe(instance)


def _assert_dns_upgrade_probe(instance: harness.Instance):
    result = (
        util.stubbornly(
            delay_s=2,
            stop=stop_after_delay(60),
            exceptions=(subprocess.CalledProcessError, subprocess.TimeoutExpired),
            reraise=True,
        )
        .on(instance)
        .exec(
            ["k8s", "kubectl", "logs", "dns-upgrade-probe", "--request-timeout=10s"],
            text=True,
            timeout=20,
        )
    )
    lines = result.stdout.splitlines()
    assert lines and all(
        line.startswith("OK ") for line in lines
    ), f"DNS probe failed: {result.stdout}"
    now = int(
        instance.exec(
            ["date", "+%s"], capture_output=True, text=True, timeout=10
        ).stdout
    )
    assert (
        now - int(lines[-1].split()[1]) < 15
    ), f"DNS probe stopped producing results: {lines[-5:]}"


def _coredns_state(instance: harness.Instance) -> dict:
    result = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "deployments,pods,hpa",
            "-n",
            "kube-system",
            "-o",
            "json",
            "--request-timeout=10s",
        ],
        capture_output=True,
        text=True,
        timeout=20,
    )
    items = json.loads(result.stdout)["items"]
    return {
        "deployment": next(
            item
            for item in items
            if item["kind"] == "Deployment" and item["metadata"]["name"] == "coredns"
        ),
        "pods": [
            item
            for item in items
            if item["kind"] == "Pod"
            and item["metadata"].get("labels", {}).get("k8s-app") == "coredns"
        ],
        "hpa": next(
            item
            for item in items
            if item["kind"] == "HorizontalPodAutoscaler"
            and item["spec"]["scaleTargetRef"]["name"] == "coredns"
        ),
    }


def _coredns_replicas_ready(state: dict, replicas: int) -> bool:
    deployment = state["deployment"]
    status = deployment.get("status", {})
    pods = state["pods"]
    return (
        deployment["spec"].get("replicas", 1) == replicas
        and status.get("observedGeneration", 0) >= deployment["metadata"]["generation"]
        and all(
            status.get(field, 0) == replicas
            for field in (
                "replicas",
                "updatedReplicas",
                "readyReplicas",
                "availableReplicas",
            )
        )
        and len(pods) == replicas
        and all(
            not pod["metadata"].get("deletionTimestamp")
            and pod["spec"].get("nodeName")
            and any(
                condition["type"] == "Ready" and condition["status"] == "True"
                for condition in pod.get("status", {}).get("conditions", [])
            )
            for pod in pods
        )
    )


def _wait_coredns_replicas(
    instance: harness.Instance, replicas: int, *, check_hpa: bool = False
) -> dict:
    state = {}

    def ready(_):
        nonlocal state
        state = _coredns_state(instance)
        hpa_status = state["hpa"].get("status", {})
        settled = _coredns_replicas_ready(state, replicas)
        if check_hpa:
            settled = settled and all(
                hpa_status.get(field) == replicas
                for field in ("currentReplicas", "desiredReplicas")
            )
        assert (
            settled
        ), f"CoreDNS has not reached {replicas} replicas: {json.dumps(state)}"
        return True

    util.stubbornly(delay_s=3, stop=stop_after_delay(180)).on(instance).until(
        ready
    ).exec(["true"], timeout=10)
    return state


def _check_coredns_hpa_scaling(
    instance: harness.Instance, instances: List[harness.Instance]
):
    state = _coredns_state(instance)
    hpa = state["hpa"]
    original_spec = hpa["spec"]
    hpa_name = hpa["metadata"]["name"]
    baseline_template = state["deployment"]["spec"]["template"]
    since = instance.exec(
        ["date", "+@%s"], capture_output=True, text=True, timeout=10
    ).stdout.strip()

    def patch_hpa(spec):
        instance.exec(
            [
                "k8s",
                "kubectl",
                "patch",
                "hpa",
                hpa_name,
                "-n",
                "kube-system",
                "--type=json",
                "-p",
                json.dumps([{"op": "replace", "path": "/spec", "value": spec}]),
                "--request-timeout=10s",
            ],
            timeout=20,
        )

    def trigger_counts():
        return [
            node.exec(
                ["journalctl", "-u", "snap.k8s.k8sd", "--since", since, "--no-pager"],
                capture_output=True,
                text=True,
                timeout=20,
            ).stdout.count("CoreDNS pods need rebalancing")
            for node in instances
        ]

    baseline_triggers = trigger_counts()
    try:
        for replicas in (len(instances) + 1, original_spec.get("minReplicas", 1)):
            LOG.info("Checking HPA-controlled CoreDNS scaling to %s replicas", replicas)
            spec = json.loads(json.dumps(original_spec))
            spec.update(minReplicas=replicas, maxReplicas=replicas)
            spec.setdefault("behavior", {}).setdefault("scaleDown", {})[
                "stabilizationWindowSeconds"
            ] = 0
            patch_hpa(spec)
            state = _wait_coredns_replicas(instance, replicas, check_hpa=True)
            assert (
                state["deployment"]["spec"]["template"] == baseline_template
            ), "Scaling changed the CoreDNS pod template"
            _assert_dns_upgrade_probe(instance)

        stable_uids = {pod["metadata"]["uid"] for pod in state["pods"]}
        until = time.monotonic() + 60
        while True:
            state = _coredns_state(instance)
            assert _coredns_replicas_ready(state, replicas), json.dumps(state)
            assert {
                pod["metadata"]["uid"] for pod in state["pods"]
            } == stable_uids, "CoreDNS pods were replaced after settling"
            assert state["deployment"]["spec"]["template"] == baseline_template
            _assert_dns_upgrade_probe(instance)
            if time.monotonic() >= until:
                break
            time.sleep(3)
        assert (
            trigger_counts() == baseline_triggers
        ), "dnsrebalancer restarted CoreDNS during scaling or observation"
    finally:
        patch_hpa(original_spec)


def _coredns_deployment(instance: harness.Instance) -> dict:
    result = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "deployment",
            "coredns",
            "-n",
            "kube-system",
            "-o",
            "json",
            "--request-timeout=10s",
        ],
        capture_output=True,
        text=True,
        timeout=20,
    )
    return json.loads(result.stdout)


def _assert_coredns_scheduling_policy(deployment: dict):
    spec = deployment["spec"]
    assert spec["strategy"]["rollingUpdate"] == {"maxSurge": 1, "maxUnavailable": 0}
    pod_spec = spec["template"]["spec"]
    preferences = pod_spec["affinity"]["podAntiAffinity"][
        "preferredDuringSchedulingIgnoredDuringExecution"
    ]
    hostname_preferences = [
        preference["podAffinityTerm"]
        for preference in preferences
        if preference["podAffinityTerm"]["topologyKey"] == "kubernetes.io/hostname"
    ]
    assert hostname_preferences, "Missing hostname anti-affinity preference"
    assert all(
        term.get("matchLabelKeys") == ["pod-template-hash"]
        for term in hostname_preferences
    )
    constraints = {
        item["topologyKey"]: item for item in pod_spec["topologySpreadConstraints"]
    }
    for topology in ("kubernetes.io/hostname", "topology.kubernetes.io/zone"):
        assert constraints[topology]["whenUnsatisfiable"] == "ScheduleAnyway"
        assert constraints[topology]["matchLabelKeys"] == ["pod-template-hash"]


def _upgrade_completed(upgrades: dict) -> bool:
    items = upgrades.get("items", [])
    assert items, f"No Upgrade resources found: {json.dumps(upgrades)}"
    upgrade = items[0]
    status = upgrade.get("status", {})
    LOG.info("Upgrade %s status: %s", upgrade["metadata"]["name"], json.dumps(status))
    assert (
        status.get("phase") == "Completed"
    ), f"Upgrade not completed: {json.dumps(upgrade)}"
    return True


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
        util.stubbornly(retries=3, delay_s=30).on(instance).exec(
            ["snap", "install", "k8s", "--classic", *util.snap_channel_args(start_snap)]
        )

    util.stubbornly(retries=3, delay_s=30).on(instance).exec(
        ["snap", "install", "k8s", "--classic", *util.snap_channel_args(start_snap)]
    )

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

        util.setup_k8s_snap(new_instance, config.SNAP)
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
        util.remove_node_with_retry(new_instance, cluster_node.id, retries=3)

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
