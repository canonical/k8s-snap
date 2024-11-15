#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
import time
from pathlib import Path
from typing import List

import pytest
from test_util import config, harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def get_gateway_service_node_port(p):
    gateway_http_port = None
    services = json.loads(p.stdout.decode())

    gateway_services = [
        svc
        for svc in services["items"]
        if (
            svc["metadata"].get("labels").get("projectcontour.io/owning-gateway-name")
            == "my-gateway"
            or svc["metadata"].get("labels").get("io.cilium.gateway/owning-gateway")
            == "my-gateway"
        )
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
    try_count = 0
    gateway_ip = None
    while gateway_ip is None and try_count < 5:
        try_count += 1
        try:
            gateway_ip = (
                instance.exec(
                    [
                        "k8s",
                        "kubectl",
                        "get",
                        "gateway",
                        "my-gateway",
                        "-o=jsonpath='{.status.addresses[0].value}'",
                    ],
                    capture_output=True,
                )
                .stdout.decode()
                .replace("'", "")
            )
        except subprocess.CalledProcessError:
            gateway_ip = None
            pass
        time.sleep(3)
    return gateway_ip


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
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

    manifest = MANIFESTS_DIR / "gateway-test.yaml"
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

    # Get gateway node port
    gateway_http_port = None
    result = (
        util.stubbornly(retries=7, delay_s=3)
        .on(instance)
        .until(lambda p: get_gateway_service_node_port(p) is not None)
        .exec(["k8s", "kubectl", "get", "service", "-o", "json"])
    )
    gateway_http_port = get_gateway_service_node_port(result)

    assert gateway_http_port is not None, "No Gateway nodePort found."

    # Test the Gateway service via loadbalancer IP.
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{gateway_http_port}"])

    gateway_ip = get_external_service_ip(instance)
    assert gateway_ip is not None, "No Gateway IP found."
    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"{gateway_ip}", "-H", "Host: foo.bar.com"])
