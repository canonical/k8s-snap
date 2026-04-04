#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from pathlib import Path
from typing import List

import pytest
from test_util import config, harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def get_gateway_service_node_port(p):
    gateway_http_port = None
    services = json.loads(p.stdout.decode())

    gateway_services = [
        svc
        for svc in services["items"]
        if svc["metadata"].get("labels", {}).get("io.cilium.gateway/owning-gateway")
        == "my-gateway"
    ]

    for svc in gateway_services:
        for port in svc["spec"]["ports"]:
            if port["port"] == 80:
                gateway_http_port = port["nodePort"]
                break
        if gateway_http_port:
            return gateway_http_port
    return None


def get_external_service_ip(instance: harness.Instance) -> str:
    """Wait for the Gateway resource to be assigned an external IP address.

    Polls the gateway status.addresses field until a non-empty IP is returned.
    Uses 60 retries with a 5-second delay (up to ~5 minutes) to account for
    slow LB-IPAM convergence in CI environments.
    """
    LOG.info("Waiting for gateway IP to be available...")

    def _has_gateway_ip(p) -> bool:
        ip = p.stdout.decode().replace("'", "").strip()
        if ip:
            return True
        LOG.info("Gateway IP not yet assigned...")
        return False

    result = (
        util.stubbornly(retries=60, delay_s=5)
        .on(instance)
        .until(_has_gateway_ip)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "gateway",
                "my-gateway",
                "-o=jsonpath='{.status.addresses[0].value}'",
            ],
        )
    )
    return result.stdout.decode().replace("'", "").strip()


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_gateway(instances: List[harness.Instance]):
    instance = instances[0]
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

    # Wait for the GatewayClass to be accepted by Cilium before applying
    # gateway resources. Without this, the Gateway controller may not be
    # ready to reconcile the Gateway and HTTPRoute resources.
    LOG.info("Waiting for GatewayClass ck-gateway to be accepted...")
    util.stubbornly(retries=30, delay_s=10).on(instance).until(
        lambda p: "True" in p.stdout.decode()
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "gatewayclass",
            "ck-gateway",
            "-o=jsonpath={.status.conditions[?(@.type=='Accepted')].status}",
        ]
    )

    manifest = MANIFESTS_DIR / "gateway-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    # Wait for the Gateway to be programmed. The Cilium gateway controller
    # needs time to reconcile the Gateway resource, create the underlying
    # service/endpoints, and set status conditions.
    LOG.info("Waiting for Gateway my-gateway to be programmed...")
    util.stubbornly(retries=30, delay_s=10).on(instance).until(
        lambda p: "True" in p.stdout.decode()
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "gateway",
            "my-gateway",
            "-o=jsonpath={.status.conditions[?(@.type=='Programmed')].status}",
        ]
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

    # Get gateway node port
    gateway_http_port = None
    result = (
        util.stubbornly(retries=20, delay_s=5)
        .on(instance)
        .until(lambda p: get_gateway_service_node_port(p))
        .exec(["k8s", "kubectl", "get", "service", "-o", "json"])
    )
    gateway_http_port = get_gateway_service_node_port(result)

    assert gateway_http_port, "No Gateway nodePort found."

    # Test the Gateway service via nodePort.
    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{gateway_http_port}"])

    gateway_ip = get_external_service_ip(instance)
    assert gateway_ip, "No Gateway IP found."
    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"{gateway_ip}", "-H", "Host: foo.bar.com"])
