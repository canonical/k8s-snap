#
# Copyright 2026 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_build(instances: List[harness.Instance]):
    """
    Test that all snap components that contain crypto functions are built dynamically
    and fail to start when FIPS is enabled on a non-compliant system.
    """
    instance = instances[0]

    if util.is_fips_enabled(instance):
        pytest.skip("Relies on a non FIPS system")

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
        "cni",
        "ctr",
        "containerd-shim-runc-v2",
    ]

    # These components should be statically built as they do not contain any crypto functions
    static_components = [
        "runc",
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

    for component in dynamic_components:
        # Verify that all components are dynamically built
        result = instance.exec(
            ["ldd", "/snap/k8s/current/bin/" + component],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stdout)
        assert (
            "not a dynamic executable" not in result.stdout
        ), f"{component} should be dynamically built"

        # Verify that the component fails if enabled on a non-FIPS system
        result = instance.exec(
            ["GOFIPS=1", "/snap/k8s/current/bin/" + component, "version"],
            capture_output=True,
            check=False,
            text=True,
        )
        LOG.info(result.stderr)
        assert result.stderr.startswith(
            "panic: opensslcrypto: FIPS mode requested (environment variable GOFIPS) but not available"
        ) or result.stderr.startswith(
            "panic: opensslcrypto: can't enable FIPS mode for OpenSSL"
        ), f"{component} should fail with FIPS enabled on non-compliant system"
