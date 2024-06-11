#
# Copyright 2024 Canonical, Ltd.
#
import logging

from test_util import harness

LOG = logging.getLogger(__name__)


def test_smoke(session_instance: harness.Instance):
    # Verify the functionality of the k8s config command during the smoke test.
    # It would be excessive to deploy a cluster solely for this purpose.
    result = session_instance.exec(
        "k8s config --server 192.168.210.41".split(), capture_output=True
    )
    config = result.stdout.decode()
    assert len(config) > 0
    assert "server: https://192.168.210.41" in config

    # Verify extra node configs
    content = session_instance.exec(
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
        args = session_instance.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()
