#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import re
import subprocess
from typing import Any, List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster status:\s*ready",
    r"control plane nodes:\s*(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})\s\(voter\)",
    r"high availability:\s*no",
    r"datastore:\s*k8s-dqlite",
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
def test_smoke(instances: List[harness.Instance]):
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
    assert (
        metadata is not None
    ), "Metadata not found in the generate-join-token response."
    assert (
        metadata.get("token") is not None
    ), "Token not found in the generate-join-token response."

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
    assert (
        metadata is not None
    ), "Metadata not found in the certificate expiry response."
    assert util.is_valid_rfc3339(
        metadata.get("expiry-date")
    ), "Token not found in the certificate expiry response."

    def status_output_matches(p: subprocess.CompletedProcess) -> bool:
        result_lines = p.stdout.decode().strip().split("\n")
        if len(result_lines) != len(STATUS_PATTERNS):
            LOG.info(
                f"wrong number of results lines, expected {len(STATUS_PATTERNS)}, got {len(result_lines)}"
            )
            return False

        for i in range(len(result_lines)):
            line, pattern = result_lines[i], STATUS_PATTERNS[i]
            if re.search(pattern, line) is None:
                LOG.info(f"could not match `{line.strip()}` with `{pattern}`")
                return False

        return True

    LOG.info("Verifying the output of `k8s status`")
    util.stubbornly(retries=15, delay_s=10).on(instance).until(
        condition=status_output_matches,
    ).exec(["k8s", "status", "--wait-ready"])


@pytest.mark.node_count(2)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-smoke.yaml").read_text()
)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_bootstrap_config(instances: List[harness.Instance]):
    """Verify that the bootstrap config does not change after changing the cluster config."""
    cp_node = instances[0]
    worker_node = instances[1]
    join_token = util.get_join_token(cp_node, worker_node, "--worker")
    util.join_cluster(
        worker_node,
        join_token,
        (config.MANIFESTS_DIR / "worker-join-smoke.yaml").read_text(),
    )

    util.wait_until_k8s_ready(cp_node, instances)
    nodes = util.ready_nodes(cp_node)
    assert len(nodes) == 2, "worker should have joined cluster"

    LOG.info("Verifying the bootstrap config does not change")
    cp_resp = get_bootstrap_config(cp_node)
    worker_resp = get_bootstrap_config(worker_node)

    toggle_ingress_enabled(cp_node)

    LOG.info(
        "Verifying the bootstrap config does not change after changing cluster config"
    )
    new_cp_resp = get_bootstrap_config(cp_node)
    new_worker_resp = get_bootstrap_config(worker_node)

    assert cp_resp["clusterConfig"] == new_cp_resp["clusterConfig"]
    assert worker_resp["clusterConfig"] == new_worker_resp["clusterConfig"]
    assert cp_resp["datastore"] == new_cp_resp["datastore"]
    assert worker_resp["datastore"] == new_worker_resp["datastore"]
    assert cp_resp["nodeTaints"] == new_cp_resp["nodeTaints"]
    assert worker_resp["nodeTaints"] == new_worker_resp["nodeTaints"]


def get_bootstrap_config(instance: harness.Instance) -> Any:
    """Get the cluster bootstrap config."""
    resp = instance.exec(
        [
            "curl",
            "-H",
            "Content-Type: application/json",
            "--unix-socket",
            "/var/snap/k8s/common/var/lib/k8sd/state/control.socket",
            "http://localhost/1.0/k8sd/cluster/config/bootstrap",
        ],
        capture_output=True,
    )
    assert resp.returncode == 0, f"Failed to get cluster bootstrap config. {resp=}"
    response = json.loads(resp.stdout.decode())
    assert (
        response["error_code"] == 0
    ), f"Failed to get cluster bootstrap config. {response=}"
    assert (
        response["error"] == ""
    ), f"Failed to get cluster bootstrap config. {response=}"

    metadata = response.get("metadata")
    assert (
        metadata is not None
    ), "Metadata not found in the cluster bootstrap config response."
    assert (
        metadata.get("clusterConfig") is not None
    ), "Config not found in the cluster bootstrap config response."
    assert (
        metadata.get("datastore") is not None
    ), "Datastore not found in the cluster bootstrap config response."
    assert (
        metadata.get("nodeTaints") is not None
    ), "Node taints not found in the cluster bootstrap config response."

    return metadata


def toggle_ingress_enabled(instance: harness.Instance) -> None:
    """Toggle the ingress enabled status and wait for the cluster to be ready."""
    resp = instance.exec("k8s get ingress.enabled".split(), capture_output=True)
    assert resp.returncode == 0, f"Failed to get ingress enabled status. {resp=}"
    is_enabled = "true" in resp.stdout.decode().strip()
    LOG.info(f"Toggling ingress enabled from {is_enabled=}")
    resp = instance.exec(f"k8s set ingress.enabled={not is_enabled}".split())
    assert resp.returncode == 0, f"Failed to toggle ingress enabled {resp=}"
    util.wait_until_k8s_ready(instance, [instance])
