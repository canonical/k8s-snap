#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
import time
from pathlib import Path
from typing import List

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


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
    while ingress_ip is None and try_count < 5:
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
                if ingress_ip is not None:
                    return ingress_ip
            except subprocess.CalledProcessError:
                ingress_ip = None
                pass
        time.sleep(3)
    return ingress_ip


def test_ingress(aio_instance: List[harness.Instance]):

    result = (
        util.stubbornly(retries=7, delay_s=3)
        .on(aio_instance)
        .until(lambda p: get_ingress_service_node_port(p) is not None)
        .exec(["k8s", "kubectl", "get", "service", "-A", "-o", "json"])
    )

    ingress_http_port = get_ingress_service_node_port(result)

    assert ingress_http_port is not None, "No ingress nodePort found."

    manifest = MANIFESTS_DIR / "ingress-test.yaml"
    aio_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(aio_instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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

    util.stubbornly(retries=5, delay_s=5).on(aio_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"])

    # Test the ingress service via loadbalancer IP
    ingress_ip = get_external_service_ip(
        aio_instance,
        [
            {"service": "ck-ingress-contour-envoy", "namespace": "projectcontour"},
            {"service": "cilium-ingress", "namespace": "kube-system"},
        ],
    )
    assert ingress_ip is not None, "No ingress IP found."
    util.stubbornly(retries=5, delay_s=5).on(aio_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"{ingress_ip}", "-H", "Host: foo.bar.com"])
