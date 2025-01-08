#
# Copyright 2025 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import List

import pytest
from test_util import harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_loadbalancer(instances: List[harness.Instance]):
    instance = instances[0]

    tester_instance = instances[1]

    instance_default_ip = util.get_default_ip(instance)
    tester_instance_default_ip = util.get_default_ip(tester_instance)

    instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)

    lb_cidr = util.find_suitable_cidr(
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

    util.stubbornly(retries=10, delay_s=2).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = (
        util.stubbornly(retries=10, delay_s=3)
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

    util.stubbornly(retries=20, delay_s=3).on(tester_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", service_ip])
