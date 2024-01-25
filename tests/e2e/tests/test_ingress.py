#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path

import pytest
from e2e_util import config, harness, util
from e2e_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def test_ingress(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    util.setup_k8s_snap(h, instance_id, snap_path)
    h.exec(instance_id, ["k8s", "bootstrap"])
    util.setup_network(h, instance_id)

    out = h.exec(
        instance_id,
        ["k8s", "enable", "ingress"],
        capture_output=True,
    )
    assert out.returncode == 0

    util.stubbornly(retries=5, delay_s=2).on(h, instance_id).until(
        lambda p: "cilium-ingress" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-n", "kube-system", "-o", "json"])

    p = h.exec(
        instance_id,
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

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
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

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
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
    h.exec(
        instance_id,
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(h, instance_id).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
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

    p = h.exec(
        instance_id,
        ["curl", f"localhost:{ingress_http_port}", "-H", "Host: foo.bar.com"],
        capture_output=True,
    )
    assert "Welcome to nginx!" in p.stdout.decode()

    h.cleanup()
