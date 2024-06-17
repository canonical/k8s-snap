#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from pathlib import Path

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


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
    util.stubbornly(retries=5, delay_s=2).on(session_instance).until(
        lambda p: "my-gateway" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = session_instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-o=json",
        ],
        capture_output=True,
    )

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

    assert (
        len(gateway_services) > 0
    ), "No gateway services found that are owned by my-gateway."

    for svc in gateway_services:
        for port in svc["spec"]["ports"]:
            if port["port"] == 80:
                gateway_http_port = port["nodePort"]
                break
        if gateway_http_port:
            break

    assert gateway_http_port is not None, "No ingress nodePort found."

    util.stubbornly(retries=5, delay_s=5).on(session_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{gateway_http_port}"])
