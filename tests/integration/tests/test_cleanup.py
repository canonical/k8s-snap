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

CONTAINERD_PATHS = [
    "/etc/containerd",
    "/run/containerd",
    "/var/lib/containerd",
]
CNI_PATH = "/opt/cni/bin"


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
def test_node_cleanup(instances: List[harness.Instance], tmp_path):
    instance = instances[0]
    util.wait_for_dns(instance)
    util.wait_for_network(instance)

    util.remove_k8s_snap(instance)

    # Check that the containerd-related folders are removed on snap removal.
    all_paths = CONTAINERD_PATHS + [CNI_PATH]
    process = instance.exec(
        ["ls", *all_paths], capture_output=True, text=True, check=False
    )
    for path in all_paths:
        assert f"cannot access '{path}': No such file or directory" in process.stderr

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

        # Check that the containerd-related folders are in the new locations.
        # If one of them is missing, this should have a non-zero exit code.
        instance.exec(["ls", *new_containerd_paths], check=True)

    for instance in instances:
        # Check that the containerd-related folders are not in the new locations after snap removal.
        util.remove_k8s_snap(instance)
        process = instance.exec(
            ["ls", *new_containerd_paths], capture_output=True, text=True, check=False
        )
        for path in new_containerd_paths:
            assert (
                f"cannot access '{path}': No such file or directory" in process.stderr
            )
