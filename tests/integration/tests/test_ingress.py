#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from pathlib import Path
from typing import List

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def check_ingress_service_and_port(p):
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


def test_ingress(session_instance: List[harness.Instance]):

    result = (
        util.stubbornly(retries=7, delay_s=3)
        .on(session_instance)
        .until(lambda p: check_ingress_service_and_port(p) is not None)
        .exec(["k8s", "kubectl", "get", "service", "-A", "-o", "json"])
    )

    ingress_http_port = check_ingress_service_and_port(result)

    assert ingress_http_port is not None, "No ingress nodePort found."

    manifest = MANIFESTS_DIR / "ingress-test.yaml"
    session_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(session_instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(session_instance).exec(
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

    util.stubbornly(retries=5, delay_s=5).on(session_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"])
