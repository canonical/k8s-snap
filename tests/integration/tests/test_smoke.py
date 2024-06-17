#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
def test_smoke(instances: List[harness.Instance]):
    instance = instances[0]

    bootstrap_smoke_config_path = "/home/ubuntu/bootstrap-smoke.yaml"
    instance.send_file(
        (config.MANIFESTS_DIR / "bootstrap-smoke.yaml").as_posix(),
        bootstrap_smoke_config_path,
    )

    instance.exec(["k8s", "bootstrap", "--file", bootstrap_smoke_config_path])
    util.wait_until_k8s_ready(instance, [instance])

    # Verify the functionality of the k8s config command during the smoke test.
    # It would be excessive to deploy a cluster solely for this purpose.
    result = instance.exec(
        "k8s config --server 192.168.210.41".split(), capture_output=True
    )
    kubeconfig = result.stdout.decode()
    assert len(kubeconfig) > 0
    assert "server: https://192.168.210.41" in kubeconfig

    # Verify extra node configs
    content = instance.exec(
        ["cat", "/var/snap/k8s/common/args/conf.d/bootstrap-extra-file.yaml"],
        capture_output=True,
    )
    assert content.stdout.decode() == "extra-args-test-file-content"

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
        args = instance.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()
