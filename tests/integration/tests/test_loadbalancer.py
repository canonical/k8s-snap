#
# Copyright 2024 Canonical, Ltd.
#
import ipaddress
import logging
import subprocess
import time
from pathlib import Path
from typing import List

import pytest
from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def find_suitable_cidr(parent_cidr: str, excluded_ips: List[str]):
    net = ipaddress.IPv4Network(parent_cidr, False)

    # Starting from the first IP address from the parent cidr,
    # we search for a /30 cidr block(4 total ips, 2 available)
    # that doesn't contain the excluded ips to avoid collisions
    # /30 because this is the smallest CIDR cilium hands out IPs from
    for i in range(4, 255, 4):
        lb_net = ipaddress.IPv4Network(f"{str(net[0]+i)}/30", False)

        contains_excluded = False
        for excluded in excluded_ips:
            if ipaddress.ip_address(excluded) in lb_net:
                contains_excluded = True
                break

        if contains_excluded:
            continue

        return str(lb_net)
    raise RuntimeError("Could not find a suitable CIDR for LoadBalancer services")


@pytest.mark.node_count(2)
def test_loadbalancer(instances: List[harness.Instance]):
    instance = instances[0]

    tester_instance = instances[1]

    instance_default_ip = util.get_default_ip(instance)
    tester_instance_default_ip = util.get_default_ip(tester_instance)

    instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)

    lb_cidr = find_suitable_cidr(
        parent_cidr=instance_default_cidr,
        excluded_ips=[instance_default_ip, tester_instance_default_ip],
    )

    instance.exec(
        ["k8s", "set", f"load-balancer.cidrs={lb_cidr}", "load-balancer.l2-mode=true"]
    )
    instance.exec(["k8s", "enable", "load-balancer"])

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    manifest = MANIFESTS_DIR / "loadbalancer-test.yaml"
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

    util.stubbornly(retries=5, delay_s=2).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = (
        util.stubbornly(retries=5, delay_s=3)
        .on(instance)
        .until(lambda p: len(p.stdout.decode().replace("'", "")) > 0)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "service",
                "my-nginx",
                "-o=jsonpath='{.status.loadBalancer.ingress[0].ip}'",
            ],
        )
    )
    service_ip = p.stdout.decode().replace("'", "")

    p = tester_instance.exec(
        ["curl", service_ip],
        capture_output=True,
    )

    assert "Welcome to nginx!" in p.stdout.decode()

    # Try to access the service via Ingress
    instance.exec(["k8s", "enable", "ingress"])
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(MANIFESTS_DIR / "ingress-test.yaml").read_bytes(),
    )
    instance.exec(
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

    try_count = 0
    ingress_ip = None
    while ingress_ip is None and try_count < 5:
        try_count += 1
        for svc in ["ck-ingress-contour-envoy", "cilium-ingress"]:
            try:
                ingress_ip = (
                    instance.exec(
                        [
                            "k8s",
                            "kubectl",
                            "--namespace",
                            "kube-system",
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
            except subprocess.CalledProcessError:
                ingress_ip = None
                pass
        time.sleep(3)

    assert ingress_ip is not None, "No ingress IP found."
    util.stubbornly(retries=5, delay_s=5).on(tester_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"{ingress_ip}", "-H", "Host: foo.bar.com"])
