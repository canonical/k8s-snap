#
# Copyright 2024 Canonical, Ltd.
#
import logging
import os
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

KUBE_CONTROLLER_MANAGER_SNAP_PORT = 10257

CONTAINERD_PATHS = [
    "/etc/containerd",
    "/run/containerd",
    "/var/lib/containerd",
]
CNI_PATH = "/opt/cni/bin"


def _assert_paths_not_exist(instance: harness.Instance, paths: List[str]):
    paths_which_exist = [
        p
        for p, exists in util.check_file_paths_exist(instance, paths).items()
        if exists
    ]
    if paths_which_exist:
        raise AssertionError(
            f"Expected the following path(s) to not exist: {paths_which_exist}"
        )


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
def test_node_cleanup(instances: List[harness.Instance], tmp_path):
    """Verifies that a `snap remove k8s` will perform proper cleanup."""
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    util.remove_k8s_snap(instance)

    # Check that the containerd-related folders are removed on snap removal.
    all_paths = CONTAINERD_PATHS + [CNI_PATH]
    _assert_paths_not_exist(instance, all_paths)

    util.setup_k8s_snap(instance, tmp_path)
    instance.exec(["k8s", "bootstrap"])


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.containerd_cfgdir("/home/ubuntu/etc/containerd")
@pytest.mark.tags(tags.NIGHTLY)
def test_node_cleanup_new_containerd_path(instances: List[harness.Instance]):
    main = instances[0]
    joiner = instances[1]

    containerd_path_bootstrap_config = (
        config.MANIFESTS_DIR / "bootstrap-containerd-path.yaml"
    ).read_text()

    main.exec(
        ["k8s", "bootstrap", "--file", "-"],
        input=str.encode(containerd_path_bootstrap_config),
    )

    join_token = util.get_join_token(main, joiner)
    joiner.exec(
        ["k8s", "join-cluster", join_token, "--file", "-"],
        input=str.encode(containerd_path_bootstrap_config),
    )

    boostrap_config = yaml.safe_load(containerd_path_bootstrap_config)
    new_containerd_paths = [
        os.path.join(
            boostrap_config["containerd-base-dir"], "k8s-containerd", p.lstrip("/")
        )
        for p in CONTAINERD_PATHS
    ]

    # /run/containerd gets created but isn't actually used (requires further
    # investigation).
    exp_missing_paths = [
        "/etc/containerd",
        "/run/containerd/containerd.sock",
        "/var/lib/containerd",
    ]

    for instance in instances:
        # Check that the containerd-related folders are not in the default locations.
        process = instance.exec(
            ["ls", *exp_missing_paths], capture_output=True, text=True, check=False
        )
        for path in exp_missing_paths:
            assert (
                f"cannot access '{path}': No such file or directory" in process.stderr
            )
        _assert_paths_not_exist(instance, exp_missing_paths)

        # Check that the containerd-related folders are in the new locations.
        # If one of them is missing, this should have a non-zero exit code.
        instance.exec(["ls", *new_containerd_paths], check=True)

    for instance in instances:
        util.remove_k8s_snap(instance)
        # Check that the containerd-related folders are not in the new locations after snap removal.
        _assert_paths_not_exist(instance, new_containerd_paths)


@pytest.mark.node_count(1)
@pytest.mark.no_setup()
@pytest.mark.tags(tags.NIGHTLY)
def test_containerd_path_cleanup_on_failed_init(
    instances: List[harness.Instance], tmp_path
):
    """Tests that a failed `bootstrap` properly cleans up any
    containerd-related paths it may have created as part of the
    failed `bootstrap`.

    It induces a bootstrap failure by pre-binding a required k8s service
    port (10257 for the kube-controller-manager) before running `k8s bootstrap`.

    NOTE: a failed `join-cluster` will trigger the exact same cleanup
    hook, so the test implicitly applies to it as well.
    """
    instance = instances[0]
    expected_code = 1
    expected_message = (
        "Encountered error(s) while verifying port availability for Kubernetes "
        "services: Port 10257 (needed by: kube-controller-manager) is already in use."
    )

    with util.open_port(KUBE_CONTROLLER_MANAGER_SNAP_PORT) as _:
        util.setup_k8s_snap(instance, tmp_path, config.SNAP, connect_interfaces=False)

        proc = instance.exec(
            ["k8s", "bootstrap"], capture_output=True, text=True, check=False
        )

        if proc.returncode != expected_code:
            raise AssertionError(
                f"Expected `k8s bootstrap` to exit with code {expected_code}, "
                f"but it exited with {proc.returncode}.\n"
                f"Stdout was: \n{proc.stdout}.\nStderr was: \n{proc.stderr}"
            )

        if expected_message not in proc.stderr:
            raise AssertionError(
                f"Expected to find port-related warning '{expected_message}' in "
                "stderr of the `k8s bootstrap` command.\n"
                f"Stdout was: \n{proc.stdout}.\nStderr was: \n{proc.stderr}"
            )

        _assert_paths_not_exist(instance, CONTAINERD_PATHS)
