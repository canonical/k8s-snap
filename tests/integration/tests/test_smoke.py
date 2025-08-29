#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster status:\s*ready",
    r"control plane nodes:\s*(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})\s\(voter\)",
    r"high availability:\s*no",
    r"network:\s*enabled",
    r"dns:\s*enabled at (\d{1,3}(?:\.\d{1,3}){3})",
    r"ingress:\s*enabled",
    r"load-balancer:\s*enabled, L2 mode",
    r"local-storage:\s*enabled at /var/snap/k8s/common/rawfile-storage",
    r"gateway\s*enabled",
]


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-smoke.yaml").read_text()
)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_smoke(instances: List[harness.Instance], datastore_type: str):
    STATUS_PATTERNS.insert(3, r"datastore:\s*{}".format(datastore_type))

    instance = instances[0]

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

    args = {
        "kube-apiserver": '--request-timeout="2m"',
        "kube-controller-manager": '--leader-elect-retry-period="3s"',
        "kube-scheduler": '--authorization-webhook-cache-authorized-ttl="11s"',
        "kube-proxy": '--config-sync-period="14m"',
        "kubelet": '--authentication-token-webhook-cache-ttl="3m"',
        "containerd": '--log-level="debug"',
    }

    if datastore_type == "etcd":
        args["etcd"] = '--log-level="info"'
    elif datastore_type == "k8s-dqlite":
        args["k8s-dqlite"] = '--watch-storage-available-size-interval="6s"'

    # For each service, verify that the extra arg was written to the args file.
    for service, value in args.items():
        args = instance.exec(
            ["cat", f"/var/snap/k8s/common/args/{service}"], capture_output=True
        )
        assert value in args.stdout.decode()

    LOG.info("Verify the functionality of the CAPI endpoints.")
    instance.exec("k8s x-capi set-auth-token my-secret-token".split())
    instance.exec("k8s x-capi set-node-token my-node-token".split())

    body = {
        "name": "my-node",
        "worker": False,
    }

    resp = instance.exec(
        [
            "curl",
            "-XPOST",
            "-H",
            "Content-Type: application/json",
            "-H",
            "capi-auth-token: my-secret-token",
            "--data",
            json.dumps(body),
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/x/capi/generate-join-token",
        ],
        capture_output=True,
    )
    response = json.loads(resp.stdout.decode())
    assert (
        response["error_code"] == 0
    ), "Failed to generate join token using CAPI endpoints."
    metadata = response.get("metadata")
    assert metadata, "Metadata not found in the generate-join-token response."
    assert metadata.get("token"), "Token not found in the generate-join-token response."

    resp = instance.exec(
        [
            "curl",
            "-XPOST",
            "-H",
            "Content-Type: application/json",
            "-H",
            "node-token: my-node-token",
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/x/capi/certificates-expiry",
        ],
        capture_output=True,
    )
    response = json.loads(resp.stdout.decode())
    assert (
        response["error_code"] == 0
    ), "Failed to get certificate expiry using CAPI endpoints."
    metadata = response.get("metadata")
    assert metadata, "Metadata not found in the certificate expiry response."
    assert util.is_valid_rfc3339(
        metadata.get("expiry-date")
    ), "Token not found in the certificate expiry response."

    LOG.info("Verifying the output of `k8s status`")
    util.stubbornly(retries=15, delay_s=10).on(instance).until(
        condition=lambda p: util.status_output_matches(p, STATUS_PATTERNS),
    ).exec(["k8s", "status", "--wait-ready"])
