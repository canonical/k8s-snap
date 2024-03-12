#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import List

from e2e_util import harness, util
from e2e_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def test_ingress(instances: List[harness.Instance]):
    instance = instances[0]

    instance.exec(["k8s", "enable", "ingress"])

    util.stubbornly(retries=5, delay_s=2).on(instance).until(
        lambda p: "cilium-ingress" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-n", "kube-system", "-o", "json"])

    p = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "kube-system",
            "cilium-ingress",
            "-o=jsonpath='{.spec.ports[?(@.name==\"http\")].nodePort}'",
        ],
        capture_output=True,
    )
    ingress_http_port = p.stdout.decode().replace("'", "")

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

    manifest = MANIFESTS_DIR / "ingress-test.yaml"
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

    p = instance.exec(
        ["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"],
        capture_output=True,
    )
    assert "Welcome to nginx!" in p.stdout.decode()
