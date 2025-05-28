#
# Copyright 2025 Canonical, Ltd.
#
import logging
import os
from typing import List

import pytest
from test_util import harness, tags

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_fips(instances: List[harness.Instance]):
    """
    Test that all snap components are built with FIPS support enabled.
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
        "runc",
    ]

    # These components should be statically built as they do not contain any crypto functions
    static_components = [
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
        assert result.returncode == 0, f"{component} should be dynamically built"
        assert "libc.so" in result.stdout, f"{component} should be dynamically built"
        assert "not a dynamic executable" not in result.stderr, f"{component} should be dynamically built"

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
        assert result.returncode != 0, f"{component} should be statically built"
        assert "not a dynamic executable" in result.stderr, f"{component} should be statically built"

    for component in dynamic_components:
        # Verify that all components are dynamically built
        result = instance.exec(
            ["ldd", "/snap/k8s/current/bin/" + component],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stdout)
        assert "not a dynamic executable" not in result.stdout, f"{component} should be dynamically built"

        # Verify that the component fails if enabled on a non-FIPS system
        result = instance.exec(
            ["GOFIPS=1", "/snap/k8s/current/bin/" + component, "version"],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stderr)
        assert "can't enable FIPS mode for OpenSSL" in result.stderr, f"{component} should fail with FIPS enabled on non-compliant system"
