#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_fips(instances: List[harness.Instance]):
    """
    Test that all snap components that contain crypto functions are built dynamically.
    """
    instance = instances[0]

    dynamic_components = [
        "k8sd",
        "kubelet",
        "kube-apiserver",
        "kube-controller-manager",
        "kube-proxy",
        "kube-scheduler",
        "kubectl",
        "k8s-apiserver-proxy",
        "containerd",
        "ctr",
    ]

    # These components should be statically built as they do not contain any crypto functions
    static_components = [
        "runc",
        "cni",
        "containerd-shim",
        "containerd-shim-runc-v1",
        "containerd-shim-runc-v2",
    ]

    for component in dynamic_components:
        # Verify that all components are dynamically built
        result = instance.exec(
            ["ldd", "/snap/k8s/current/bin/" + component],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stdout)
        LOG.info(result.stderr)
        assert "libc.so" in result.stdout, f"{component} should be dynamically built"
        assert (
            "not a dynamic executable" not in result.stderr
        ), f"{component} should be dynamically built"

    for component in static_components:
        # Verify that all components are statically built
        result = instance.exec(
            ["ldd", "/snap/k8s/current/bin/" + component],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stdout)
        LOG.info(result.stderr)
        assert (
            "not a dynamic executable" in result.stderr
            or "statically linked" in result.stdout
        ), f"{component} should be statically built"
