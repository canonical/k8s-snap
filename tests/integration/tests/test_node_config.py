#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
@pytest.mark.disable_k8s_bootstrapping()
def test_extra_node_args(instances: List[harness.Instance]):
    cluster_node = instances[0]
    joining_cp_node = instances[1]
    joining_worker_node = instances[2]

    extra_args_config_path = "/home/ubuntu/extra-args.yaml"
    extra_args_test_file_path = "/home/ubuntu/extra-args-test-file.yaml"

    cluster_node.send_file(
        (config.MANIFESTS_DIR / "bootstrap-extra-node-args.yaml").as_posix(),
        extra_args_config_path,
    )
    # Create a test file on the instance that is loaded by the extra args config.
    cluster_node.exec(["touch", extra_args_test_file_path])
    cluster_node.exec(["k8s", "bootstrap", "--file", extra_args_config_path])

    # Check if the extra file was written to the correct location.
    cluster_node.exec(
        ["cat", "/var/snap/k8s/common/args/conf.d/bootstrap-extra-file.yaml"]
    )

    # For each service, verify that the extra arg was written to the args file.
    for service, value in {
        "kube-apiserver": "--request-timeout=2m",
        "kube-controller-manager": "--leader-elect-retry-period=3s",
        "kube-scheduler": "--authorization-webhook-cache-authorized-ttl=11s",
        "kube-proxy": "--config-sync-period=14m",
        "kubelet": "--authentication-token-webhook-cache-ttl=3m",
        "containerd": "--log-level=debug",
        "k8s-dqlite": "--watch-storage-available-size-interval=6s",
    }.items():
        args = cluster_node.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()

    # Join a control-plane to the cluster.
    joining_cp_node.send_file(
        (config.MANIFESTS_DIR / "join-extra-node-args.yaml").as_posix(),
        extra_args_config_path,
    )
    joining_cp_node.exec(["touch", extra_args_test_file_path])

    join_token = util.get_join_token(cluster_node, joining_cp_node)
    util.join_cluster(joining_cp_node, join_token, "--file", extra_args_config_path)

    joining_cp_node.exec(
        ["cat", "/var/snap/k8s/common/args/conf.d/join-extra-file.yaml"]
    )

    for service, value in {
        "kube-apiserver": "--request-timeout=3m",
        "kube-controller-manager": "--leader-elect-retry-period=4s",
        "kube-scheduler": "--authorization-webhook-cache-authorized-ttl=12s",
        "kube-proxy": "--config-sync-period=13m",
        "kubelet": "--authentication-token-webhook-cache-ttl=4m",
        "containerd": "--log-level=warning",
        "k8s-dqlite": "--watch-storage-available-size-interval=7s",
    }.items():
        args = joining_cp_node.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()

    # Join a worker to the cluster.
    joining_worker_node.send_file(
        (config.MANIFESTS_DIR / "worker-extra-node-args.yaml").as_posix(),
        extra_args_config_path,
    )
    joining_worker_node.exec(["touch", extra_args_test_file_path])

    join_token = util.get_join_token(cluster_node, joining_worker_node, "--worker")
    util.join_cluster(joining_worker_node, join_token, "--file", extra_args_config_path)

    joining_worker_node.exec(
        ["cat", "/var/snap/k8s/common/args/conf.d/worker-extra-file.yaml"]
    )

    for service, value in {
        "kube-proxy": "--config-sync-period=12m",
        "kubelet": "--authentication-token-webhook-cache-ttl=5m",
        "containerd": "--log-level=error",
        "k8s-apiserver-proxy": "--refresh-interval=29s",
    }.items():
        args = joining_worker_node.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()
