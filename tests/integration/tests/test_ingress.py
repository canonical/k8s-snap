#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import List

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def test_ingress(instance: List[harness.Instance]):

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

    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"])
