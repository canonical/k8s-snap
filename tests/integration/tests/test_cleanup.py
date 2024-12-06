#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

KUBE_CONTROLLER_MANAGER_SNAP_PORT = 10257

CONTAINERD_PATHS = [
    "/etc/containerd",
    "/opt/cni/bin",
    "/run/containerd",
    "/var/lib/containerd",
]


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
    _assert_paths_not_exist(instance, CONTAINERD_PATHS)

    util.setup_k8s_snap(instance, tmp_path)
    instance.exec(["k8s", "bootstrap"])


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
