#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.skipif(
    util.host_is_fips_enabled(),
    reason="Skip on FIPS systems since we use the convenient early FIPS failures for this tests.",
)
@pytest.mark.tags(tags.NIGHTLY)
def test_user_config(instances: List[harness.Instance]):
    """Verifies that the snap and services environment variables set by the user are loaded correctly.

    Checks that:
    - Environment variables are properly passed to commands.
    - `snap-env` is passed to all services.
    - `<service-env>` is passed to the service.
    """
    instance = instances[0]

    # Verify that environment variables are properly passed to commands.
    result = instance.exec(
        "GOFIPS=1 k8s status".split(), capture_output=True, check=False
    )
    assert result.returncode != 0
    assert (
        "Please run this service on a FIPS-enabled host to use FIPS mode."
        in result.stdout.decode()
    )

    # Write a snap-env file which should be read by all services.
    instance.exec("echo 'GOFIPS=1' > /var/snap/k8s/common/args/snap-env".split())

    # Verify that the snap-env file is read by multiple services.
    result = instance.exec("k8s status".split(), capture_output=True, check=False)
    assert result.returncode != 0
    assert (
        "Please run this service on a FIPS-enabled host to use FIPS mode."
        in result.stdout.decode()
    )

    # Fails because of FIPS mode.
    result = instance.exec("snap start k8s.containerd".split(), check=False)
    assert result.returncode != 0

    instance.exec("rm /var/snap/k8s/common/args/snap-env".split())

    # Write a service-env file which should be read by the k8s service.
    instance.exec("echo 'GOFIPS=1' > /var/snap/k8s/common/args/k8s-env".split())
    result = instance.exec("k8s status".split(), capture_output=True, check=False)
    assert result.returncode != 0
    assert (
        "Please run this service on a FIPS-enabled host to use FIPS mode."
        in result.stdout.decode()
    )
    instance.exec("rm /var/snap/k8s/common/args/k8s-env".split())
