#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import time
from typing import List

import pytest
from test_util import config, harness, k8s, tags, util

LOG = logging.getLogger(__name__)


def check_pod_for_fips_error(
    instance: harness.Instance,
    pod_name: str,
    namespace: str,
) -> bool:
    """
    Check if a pod has FIPS-related errors in logs or status.

    Args:
        instance: instance on which to execute check
        pod_name: name of the pod to check
        namespace: namespace of the pod

    Returns:
        True if FIPS error found, False otherwise
    """
    patterns = ("FIPS mode requested", "FIPS mode", "opensslcrypto")

    def _check_logs(previous: bool = False) -> bool:
        cmd = ["k8s", "kubectl", "logs", pod_name, "-n", namespace, "--tail=50"]
        if previous:
            cmd.append("--previous")
        result = instance.exec(cmd, capture_output=True, check=False, text=True)
        if result.returncode != 0:
            return False
        logs = result.stdout or ""
        if any(pat in logs for pat in patterns):
            LOG.info(
                f"Found FIPS error in {'previous ' if previous else ''}logs for pod {pod_name}"
            )
            return True
        return False

    return _check_logs(previous=False) or _check_logs(previous=True)


def verify_resource_fips_failure(
    instance: harness.Instance,
    namespace: str,
    resource_type: str,
    name: str,
    max_retries: int = 5,
    retry_delay_s: int = 10,
) -> bool:
    """
    Verify that a resource's pods fail with FIPS errors after patching.

    Args:
        instance: instance on which to execute check
        namespace: namespace of the resource
        resource_type: type of resource
        name: name of the resource
        max_retries: maximum number of retries
        retry_delay_s: delay between retries in seconds

    Returns:
        True if FIPS error found, False otherwise
    """
    for attempt in range(max_retries):
        LOG.info(f"Attempt {attempt + 1} to find FIPS errors in pods")

        # Get all pods in namespace
        result = instance.exec(
            [
                "k8s",
                "kubectl",
                "get",
                "pods",
                "-n",
                namespace,
                "-o",
                "json",
            ],
            capture_output=True,
            text=True,
        )

        pods_data = json.loads(result.stdout)

        # Check for pods that match this resource
        for pod in pods_data.get("items", []):
            pod_name = pod["metadata"]["name"]

            # Check if this pod belongs to our resource
            owner_refs = pod["metadata"].get("ownerReferences", [])
            is_owned = False
            for owner in owner_refs:
                if (
                    resource_type == "daemonset"
                    and owner.get("kind") == "DaemonSet"
                    and owner.get("name") == name
                ):
                    is_owned = True
                    break
                elif (
                    resource_type == "deployment" and owner.get("kind") == "ReplicaSet"
                ):
                    rs_name = owner.get("name")
                    if rs_name.startswith(name + "-"):
                        is_owned = True
                        break

            if not is_owned:
                continue

            LOG.info(f"Checking pod {pod_name} for FIPS errors")

            # Check pod logs for FIPS error
            if check_pod_for_fips_error(instance, pod_name, namespace):
                return True

        # Wait before retrying
        if attempt < max_retries - 1:
            LOG.info("No FIPS errors found yet, waiting...")
            time.sleep(retry_delay_s)

    return False


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.NIGHTLY)
def test_fips_images(instances: List[harness.Instance]):
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
    resources = k8s.get_workload_resources_in_namespaces(
        instance,
        namespaces_to_check,
        resource_types,
        # Exclude local-storage because it's not implemented in Go.
        exclude=["*ck-storage*"],
    )

    assert len(resources) > 0, "No resources found in the specified namespaces"
    LOG.info(f"Found {len(resources)} resources to test")

    # For each resource, patch it to add GOFIPS=1 and verify it fails
    for resource in resources:
        namespace = resource.namespace
        resource_type = resource.type
        name = resource.name

        LOG.info(f"Testing FIPS compliance for {resource_type}/{name} in {namespace}")

        k8s.update_resource_container_env(
            instance, namespace, resource_type, name, {"GOFIPS": "1"}
        )

        # Wait for pods to restart with GOFIPS=1
        LOG.info(f"Waiting for {resource_type}/{name} pods to restart with GOFIPS=1...")
        found_fips_error = verify_resource_fips_failure(
            instance, namespace, resource_type, name
        )

        assert found_fips_error, (
            f"Expected pods for {resource_type}/{name} in {namespace} to fail with FIPS error, "
            "but no FIPS-related errors were found"
        )

        LOG.info(f"Verified FIPS error for {resource_type}/{name}")

        LOG.info(f"Restoring {resource_type}/{name} to GOFIPS=0")
        k8s.update_resource_container_env(
            instance, namespace, resource_type, name, {"GOFIPS": "0"}
        )
        LOG.info(f"Waiting for {resource_type}/{name} pods to restart with GOFIPS=0...")
        util.stubbornly().until(
            lambda _: k8s.resource_ready(instance, namespace, resource_type, name)
        ).exec(["echo", "waiting..."])

    LOG.info("All container images successfully verified as FIPS-compiled")
