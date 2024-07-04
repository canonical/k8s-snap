#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from pathlib import Path

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)

def check_gateway_service_and_port(p):
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
            print(f"Found gateway service with nodePort: {gateway_http_port}") #TODO: remove print
            return gateway_http_port
    return None

def test_gateway(session_instance: harness.Instance):
    manifest = MANIFESTS_DIR / "gateway-test.yaml"
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

    # Get gateway node port
    gateway_http_port = None
    gateway_http_port = util.stubbornly(retries=7, delay_s=3).on(session_instance).until(
        lambda p: check_gateway_service_and_port(p) is not None
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    assert gateway_http_port is not None, "No ingress nodePort found."

    util.stubbornly(retries=5, delay_s=5).on(session_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{gateway_http_port}"])
