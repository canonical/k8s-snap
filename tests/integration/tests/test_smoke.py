#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import re
import subprocess
from typing import List
import time
from pathlib import Path

import pytest
from test_util import config, harness, tags, util
from test_util.config import MANIFESTS_DIR

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
        "kube-apiserver": '--request-timeout="2m"',
        "kube-controller-manager": '--leader-elect-retry-period="3s"',
        "kube-scheduler": '--authorization-webhook-cache-authorized-ttl="11s"',
        "kube-proxy": '--config-sync-period="14m"',
        "kubelet": '--authentication-token-webhook-cache-ttl="3m"',
        "containerd": '--log-level="debug"',
        "k8s-dqlite": '--watch-storage-available-size-interval="6s"',
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

    def status_output_matches(p: subprocess.CompletedProcess) -> bool:
        result_lines = p.stdout.decode().strip().split("\n")
        if len(result_lines) != len(STATUS_PATTERNS):
            LOG.info(
                f"wrong number of results lines, expected {len(STATUS_PATTERNS)}, got {len(result_lines)}"
            )
            return False

        for i in range(len(result_lines)):
            line, pattern = result_lines[i], STATUS_PATTERNS[i]
            if not re.search(pattern, line):
                LOG.info(f"could not match `{line.strip()}` with `{pattern}`")
                return False

        return True

    # NOTE(Hue): Making Ingress Great Again
    instance_default_ip = util.get_default_ip(instance)
    instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)
    lb_cidr = util.find_suitable_cidr(
        parent_cidr=instance_default_cidr,
        excluded_ips=[instance_default_ip],
    )
    instance.exec(
        ["k8s", "set", f"load-balancer.cidrs={lb_cidr}", "load-balancer.l2-mode=true"]
    )
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    result = (
        util.stubbornly(retries=20, delay_s=3)
        .on(instance)
        .until(lambda p: get_ingress_service_node_port(p))
        .exec(["k8s", "kubectl", "get", "service", "-A", "-o", "json"])
    )

    ingress_http_port = get_ingress_service_node_port(result)

    assert ingress_http_port, "No ingress nodePort found."

    manifest = MANIFESTS_DIR / "ingress-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=my-nginx",
            "--timeout",
            "180s",
        ]
    )

    util.stubbornly(retries=10, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"])

    # Test the ingress service via loadbalancer IP
    ingress_ip = get_external_service_ip(
        instance,
        [
            {"service": "ck-ingress-contour-envoy", "namespace": "projectcontour"},
            {"service": "cilium-ingress", "namespace": "kube-system"},
        ],
    )
    assert ingress_ip, "No ingress IP found."
    util.stubbornly(retries=10, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"{ingress_ip}", "-H", "Host: foo.bar.com"])

    LOG.info("Verifying the output of `k8s status`")
    util.stubbornly(retries=15, delay_s=10).on(instance).until(
        condition=status_output_matches,
    ).exec(["k8s", "status", "--wait-ready"])


def get_ingress_service_node_port(p):
    ingress_http_port = None
    services = json.loads(p.stdout.decode())

    ingress_services = [
        svc
        for svc in services["items"]
        if (
            svc["metadata"]["name"] == "ck-ingress-contour-envoy"
            or svc["metadata"]["name"] == "cilium-ingress"
        )
    ]

    for svc in ingress_services:
        for port in svc["spec"]["ports"]:
            if port["port"] == 80:
                ingress_http_port = port["nodePort"]
                break
        if ingress_http_port:
            return ingress_http_port
    return None


def get_external_service_ip(instance: harness.Instance, service_namespace) -> str:
    try_count = 0
    ingress_ip = None
    while not ingress_ip and try_count < 60:
        try_count += 1
        for svcns in service_namespace:
            svc = svcns["service"]
            namespace = svcns["namespace"]
            try:
                ingress_ip = (
                    instance.exec(
                        [
                            "k8s",
                            "kubectl",
                            "--namespace",
                            namespace,
                            "get",
                            "service",
                            svc,
                            "-o=jsonpath='{.status.loadBalancer.ingress[0].ip}'",
                        ],
                        capture_output=True,
                    )
                    .stdout.decode()
                    .replace("'", "")
                )
                if ingress_ip:
                    return ingress_ip
            except subprocess.CalledProcessError:
                ingress_ip = None
                pass
        time.sleep(3)
    return ingress_ip
