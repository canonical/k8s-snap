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


def test_ingress(session_instance: List[harness.Instance]):

    util.stubbornly(retries=5, delay_s=2).on(session_instance).until(
        lambda p: "ingress" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-n", "kube-system", "-o", "json"])

    p = session_instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-A",
            "-o=json",
        ],
        capture_output=True,
    )

    ingress_http_port = None
    services = json.loads(p.stdout.decode())
    for svc in services["items"]:
        if "ingress" in svc["metadata"]["name"]:
            LOG.info(f"Found service {svc['metadata']['name']}")
            for port in svc["spec"]["ports"]:
                if port["port"] == 80:
                    ingress_http_port = port.get("nodePort")
                    break
            if ingress_http_port:
                break

    assert ingress_http_port is not None, "No ingress nodePort found."
    LOG.info(f"Found http port: {ingress_http_port}")

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
