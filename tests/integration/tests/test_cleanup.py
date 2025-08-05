#
# Copyright 2025 Canonical, Ltd.
#
import logging
import os
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

CONTAINERD_SOCKET_DIRECTORY_CLASSIC = "/run/containerd"

CONTAINERD_PATHS = [
    "/etc/containerd",
    CONTAINERD_SOCKET_DIRECTORY_CLASSIC,
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
def test_node_cleanup(instances: List[harness.Instance], tmp_path, datastore_type: str):
    """Verifies that a `snap remove k8s` will perform proper cleanup."""
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    util.remove_k8s_snap(instance)

    # Check that the containerd-related folders are removed on snap removal.
    all_paths = CONTAINERD_PATHS + [CNI_PATH]
    _assert_paths_not_exist(instance, all_paths)

    util.setup_k8s_snap(instance)
    util.bootstrap(instance, datastore_type=datastore_type)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.containerd_cfgdir("/home/ubuntu/k8s-containerd/etc/containerd")
@pytest.mark.tags(tags.NIGHTLY)
def test_node_cleanup_new_containerd_path(
    instances: List[harness.Instance], datastore_type: str
):
    main = instances[0]
    joiner = instances[1]

    containerd_path_bootstrap_config = (
        config.MANIFESTS_DIR / "bootstrap-containerd-path.yaml"
    ).read_text()
    containerd_path_join_config = """
containerd-base-dir: /home/ubuntu
"""

    util.bootstrap(
        main,
        datastore_type=datastore_type,
        bootstrap_config=containerd_path_bootstrap_config,
    )

    join_token = util.get_join_token(main, joiner)
    joiner.exec(
        ["k8s", "join-cluster", join_token, "--file", "-"],
        input=str.encode(containerd_path_join_config),
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

    # Ensure that the cluster actually becomes available.
    util.wait_until_k8s_ready(main, instances)

    for instance in instances:
        util.remove_k8s_snap(instance)
        # Check that the containerd-related folders are not in the new locations after snap removal.
        _assert_paths_not_exist(instance, new_containerd_paths)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_containerd_path_cleanup_on_failed_init(
    instances: List[harness.Instance], datastore_type: str
):
    """Tests that a failed `bootstrap` properly cleans up any
    containerd-related paths it may have created as part of the
    failed `bootstrap`.

    It introduces a bootstrap failure by supplying an incorrect argument to the kube-apiserver.

    The bootstrap/join-cluster aborting behavior was added in this PR:
    https://github.com/canonical/k8s-snap/pull/772

    NOTE: a failed `join-cluster` will trigger the exact same cleanup
    hook, so the test implicitly applies to it as well.
    """
    instance = instances[0]
    expected_code = 1

    fail_bootstrap_config = (config.MANIFESTS_DIR / "bootstrap-fail.yaml").read_text()

    proc = util.bootstrap(
        instance,
        datastore_type=datastore_type,
        bootstrap_config=fail_bootstrap_config,
        check=False,
    )

    if proc.returncode != expected_code:
        raise AssertionError(
            f"Expected `k8s bootstrap` to exit with code {expected_code}, "
            f"but it exited with {proc.returncode}.\n"
        )

    _assert_paths_not_exist(instance, CONTAINERD_PATHS)
