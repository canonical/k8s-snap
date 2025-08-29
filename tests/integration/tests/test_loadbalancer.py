#
# Copyright 2025 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import List

import pytest
from test_util import config, harness, tags, util
from test_util.config import MANIFESTS_DIR, SUBSTRATE

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
# For loadbalancer communication
@pytest.mark.required_ports(80)
def test_loadbalancer_ipv4(
    instances: List[harness.Instance], cluster_network_type: str
):
    _test_loadbalancer(instances, cluster_network_type)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-ipv6-only.yaml").read_text()
)
@pytest.mark.infra_network_type("Dualstack")
@pytest.mark.cluster_network_type("IPv6")
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_only(
    instances: List[harness.Instance], cluster_network_type: str
):
    _test_loadbalancer(instances, cluster_network_type)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
)
@pytest.mark.infra_network_type("Dualstack")
@pytest.mark.cluster_network_type("Dualstack")
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_dualstack(
    instances: List[harness.Instance], cluster_network_type: str
):
    _test_loadbalancer(instances, cluster_network_type)


def _test_loadbalancer(instances: List[harness.Instance], network_type: str):
    instance = instances[0]
    tester_instance = instances[1]

    lb_cidrs = []

    def get_lb_cidr(ipv6_cidr: bool):
        instance_default_ip = util.get_default_ip(instance, ipv6=ipv6_cidr)
        tester_instance_default_ip = util.get_default_ip(
            tester_instance, ipv6=ipv6_cidr
        )
        instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)
        lb_cidr = util.find_suitable_cidr(
            parent_cidr=instance_default_cidr,
            excluded_ips=[instance_default_ip, tester_instance_default_ip],
        )
        return lb_cidr

    if network_type in ("IPv4", "Dualstack"):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=False))
    if network_type in ("IPv6", "Dualstack"):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=True))
    lb_cidr_str = ",".join(lb_cidrs)

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)
    instance.exec(
        [
            "k8s",
            "set",
            f"load-balancer.cidrs={lb_cidr_str}",
            "load-balancer.l2-mode=true",
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

    util.stubbornly(retries=3, delay_s=5).on(instance).exec(
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
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = (
        util.stubbornly(retries=10, delay_s=5)
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
    if ":" in service_ip:
        service_ip = "[" + service_ip + "]"

    LOG.info(f"Reaching out to service with service_ip = {service_ip}")
    util.stubbornly(retries=40, delay_s=5).on(tester_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", service_ip])
