#
# Copyright 2025 Canonical, Ltd.
#
import logging
import time
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.NIGHTLY)
def test_fips_rocks(instances: List[harness.Instance]):
    """
    Test that all container images are FIPS-compiled by verifying they fail to start
    when GOFIPS=1 is set on a non-FIPS system.

    This test:
    1. Bootstraps a Canonical Kubernetes cluster normally
    2. Identifies all deployments and daemonsets in kube-system and metallb-system
    3. Sequentially patches each to add GOFIPS=1 environment variable to all containers
    4. Verifies that pods fail to start with expected FIPS panic message
    """
    instance = instances[0]

    if util.is_fips_enabled(instance):
        pytest.skip("Test requires a non-FIPS system")

    # Wait for cluster to be ready
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    # Define namespaces and resource types to check
    namespaces_to_check = ["kube-system", "metallb-system"]
    resource_types = ["daemonset", "deployment"]

    # Collect all resources in the specified namespaces
    resources = util.get_resources_in_namespaces(
        instance, namespaces_to_check, resource_types
    )

    assert len(resources) > 0, "No resources found in the specified namespaces"
    LOG.info(f"Found {len(resources)} resources to test")

    # For each resource, patch it to add GOFIPS=1 and verify it fails
    for resource in resources:
        namespace = resource["namespace"]
        resource_type = resource["type"]
        name = resource["name"]

        LOG.info(f"Testing FIPS compliance for {resource_type}/{name} in {namespace}")

        # Patch the resource to add GOFIPS=1
        if not util.patch_resource_with_gofips(
            instance, namespace, resource_type, name
        ):
            continue

        # Wait for pods to restart with GOFIPS=1
        LOG.info(f"Waiting for {resource_type}/{name} pods to restart with GOFIPS=1...")
        time.sleep(10)

        # Verify that pods fail with FIPS errors
        found_fips_error = util.verify_resource_fips_failure(
            instance, namespace, resource_type, name
        )

        assert found_fips_error, (
            f"Expected pods for {resource_type}/{name} in {namespace} to fail with FIPS error, "
            "but no FIPS-related errors were found"
        )

        LOG.info(f"Verified FIPS error for {resource_type}/{name}")

    LOG.info("All container images successfully verified as FIPS-compiled")
