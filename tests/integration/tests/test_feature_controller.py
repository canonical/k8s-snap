#
# Copyright 2026 Canonical, Ltd.
#
import logging
import re
import subprocess
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster:\s*ready",
    r"nodes:\s*\d+\s*control-plane",
    r"^\s*$",
    r"Networking:",
    r"●\s*network\s*\(cilium",
    r"\s+\S.*",
    r"^\s*$",
    r"●\s*dns\s*\(coredns",
    r"\s+\S.*",
    r"^\s*$",
    r"●\s*load-balancer\s*\(metallb",
    r"\s+\S.*",
    r"^\s*$",
    r"●\s*ingress\s*\(cilium",
    r"\s+\S.*",
    r"^\s*$",
    r"●\s*gateway\s*\(cilium",
    r"\s+enabled",
    r"^\s*$",
    r"Storage:",
    r"●\s*local-storage\s*\(rawfile-csi",
    r"\s+\S.*",
    r"^\s*$",
    r"Observability:",
    r"●\s*metrics-server\s*\(metrics-server",
    r"\s+\S.*",
    r"^\s*$",
    r"Suggestions:",
    r"k8s kubectl get nodes\s+View detailed node information",
    r"k8s get\s+View cluster configuration",
]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_feature_controller(instances: List[harness.Instance]):
    """
    Verifies that the feature controller won't get stuck in a chaotic situation.
    """

    instance = instances[0]

    for _ in range(20):
        instance.exec(
            [
                "k8s",
                "disable",
                "dns",
                "gateway",
                "ingress",
                "load-balancer",
                "metrics-server",
                "local-storage",
            ]
        )
        instance.exec(
            [
                "pkill",
                "-f",
                "-9",
                "/snap/k8s/current/bin/kube-apiserver",
            ],
            check=False,
        )
        instance.exec(
            [
                "k8s",
                "enable",
                "dns",
                "gateway",
                "ingress",
                "load-balancer",
                "metrics-server",
                "local-storage",
            ]
        )
        instance.exec(
            [
                "pkill",
                "-f",
                "-9",
                "/snap/k8s/current/bin/kube-apiserver",
            ],
            check=False,
        )

    def status_output_matches(p: subprocess.CompletedProcess) -> bool:
        result_lines = p.stdout.decode().strip().split("\n")
        if len(result_lines) != len(STATUS_PATTERNS):
            LOG.info(
                f"wrong number of results lines, expected {len(STATUS_PATTERNS)}, got {len(result_lines)}"
            )
            return False

        for i in range(len(result_lines)):
            line, pattern = result_lines[i], STATUS_PATTERNS[i]
            if not re.search(pattern, line):
                LOG.info(f"could not match `{line.strip()}` with `{pattern}`")
                return False

        return True

    LOG.info("Verifying the output of `k8s status`")
    util.stubbornly(retries=30, delay_s=20).on(instance).until(
        condition=status_output_matches,
    ).exec(["k8s", "status", "--wait-ready"])
