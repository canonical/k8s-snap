#
# Copyright 2024 Canonical, Ltd.
#
import ipaddress
import logging
from pathlib import Path
from typing import List

import pytest
from e2e_util import harness, util
from e2e_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def get_default_cidr(instance: harness.Instance, instance_default_ip: str):
    # ----
    # 1:  lo    inet 127.0.0.1/8 scope host lo .....
    # 28: eth0  inet 10.42.254.197/24 metric 100 brd 10.42.254.255 scope global dynamic eth0 ....
    # ----
    # Fetching the cidr for the default interface by matching with instance ip from the output
    p = instance.exec(["ip", "-o", "-f", "inet", "addr", "show"], capture_output=True)
    out = p.stdout.decode().split(" ")
    return [i for i in out if instance_default_ip in i][0]


def get_default_ip(instance: harness.Instance):
    # ---
    # default via 10.42.254.1 dev eth0 proto dhcp src 10.42.254.197 metric 100
    # ---
    # Fetching the default IP address from the output, e.g. 10.42.254.197
    p = instance.exec(
        ["ip", "-o", "-4", "route", "show", "to", "default"], capture_output=True
    )
    return p.stdout.decode().split(" ")[8]


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

    instance_default_ip = get_default_ip(instance)
    tester_instance_default_ip = get_default_ip(tester_instance)

    instance_default_cidr = get_default_cidr(instance, instance_default_ip)

    lb_cidr = find_suitable_cidr(
        parent_cidr=instance_default_cidr,
        excluded_ips=[instance_default_ip, tester_instance_default_ip],
    )

    instance.exec(
        ["k8s", "set", f"load-balancer.cidrs={lb_cidr}", "load-balancer.l2-mode=true"]
    )
    instance.exec(["k8s", "enable", "load-balancer"])

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "io.cilium/app=operator",
            "--timeout",
            "180s",
        ]
    )

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=cilium",
            "--timeout",
            "180s",
        ]
    )

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
