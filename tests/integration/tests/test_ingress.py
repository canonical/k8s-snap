#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
import subprocess
import time
from typing import List

import pytest
from test_util import config, harness, tags, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def get_ingress_class_and_service_node_port(p):
    ingress_http_port = None
    services = json.loads(p.stdout.decode())

    ingress_class = None
    ingress_services = []
    for svc in services["items"]:
        if svc["metadata"]["name"] == "ck-ingress-contour-envoy":
            ingress_class = "ck-ingress"
            ingress_services.append(svc)
        elif svc["metadata"]["name"] == "cilium-ingress":
            ingress_class = "cilium"
            ingress_services.append(svc)

    for svc in ingress_services:
        for port in svc["spec"]["ports"]:
            if port["port"] == 80:
                ingress_http_port = port["nodePort"]
                break
        if ingress_http_port:
            return ingress_http_port, ingress_class
    return None, ingress_class


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


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_ingress(instances: List[harness.Instance]):
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

    def has_ingress_node_port(p: subprocess.CompletedProcess) -> bool:
        port, _ = get_ingress_class_and_service_node_port(p)
        return bool(port)

    result = (
        util.stubbornly(retries=20, delay_s=3)
        .on(instance)
        .until(has_ingress_node_port)
        .exec(["k8s", "kubectl", "get", "service", "-A", "-o", "json"])
    )

    ingress_http_port, ingress_class = get_ingress_class_and_service_node_port(result)

    assert ingress_http_port, "No ingress nodePort found."
    assert ingress_class, "No ingress class found."

    manifest_tmpl_path = MANIFESTS_DIR / "ingress-test.yaml.tmpl"
    manifest = manifest_tmpl_path.read_text().format(ingress_class_name=ingress_class)

    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.encode("utf-8"),
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
