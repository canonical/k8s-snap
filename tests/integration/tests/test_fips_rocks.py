#
# Copyright 2025 Canonical, Ltd.
#
import json
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

    # Define namespaces and resource types to check for ROCK images
    namespaces_to_check = ["kube-system", "metallb-system"]
    resource_types = ["daemonset", "deployment"]

    # Collect all resources in the specified namespaces
    rock_resources = []

    for namespace in namespaces_to_check:
        for resource_type in resource_types:
            LOG.info(f"Checking {resource_type}s in namespace {namespace}")
            result = instance.exec(
                [
                    "k8s",
                    "kubectl",
                    "get",
                    resource_type,
                    "-n",
                    namespace,
                    "-o",
                    "json",
                ],
                capture_output=True,
                check=False,
                text=True,
            )

            if result.returncode != 0:
                LOG.info(
                    f"No {resource_type}s found in namespace {namespace} or namespace doesn't exist"
                )
                continue

            resources = json.loads(result.stdout)

            for item in resources.get("items", []):
                name = item["metadata"]["name"]
                # Check all containers in the resource
                containers = item["spec"]["template"]["spec"].get("containers", [])

                # Include all resources that have containers
                if containers:
                    rock_resources.append(
                        {
                            "namespace": namespace,
                            "type": resource_type,
                            "name": name,
                        }
                    )
                    LOG.info(f"Found resource: {resource_type}/{name} in {namespace}")

    assert len(rock_resources) > 0, "No resources found in the specified namespaces"

    LOG.info(f"Found {len(rock_resources)} resources to test")

    # For each resource, patch it to add GOFIPS=1 and verify it fails
    for resource in rock_resources:
        namespace = resource["namespace"]
        resource_type = resource["type"]
        name = resource["name"]

        LOG.info(f"Testing FIPS compliance for {resource_type}/{name} in {namespace}")

        # Get the current resource definition
        result = instance.exec(
            [
                "k8s",
                "kubectl",
                "get",
                resource_type,
                name,
                "-n",
                namespace,
                "-o",
                "json",
            ],
            capture_output=True,
            text=True,
        )
        resource_def = json.loads(result.stdout)

        # Prepare patch to add GOFIPS=1 to all containers
        containers = resource_def["spec"]["template"]["spec"]["containers"]

        # Build JSON patch for each container
        patches = []
        for i, container in enumerate(containers):
            # Patch all containers to add GOFIPS=1
            env = container.get("env", [])
            # Check if GOFIPS already exists
            gofips_exists = any(e.get("name") == "GOFIPS" for e in env)

            if gofips_exists:
                # Update existing GOFIPS value
                for j, e in enumerate(env):
                    if e.get("name") == "GOFIPS":
                        patches.append(
                            {
                                "op": "replace",
                                "path": f"/spec/template/spec/containers/{i}/env/{j}/value",
                                "value": "1",
                            }
                        )
            else:
                # Add GOFIPS to env
                if env:
                    patches.append(
                        {
                            "op": "add",
                            "path": f"/spec/template/spec/containers/{i}/env/-",
                            "value": {"name": "GOFIPS", "value": "1"},
                        }
                    )
                else:
                    patches.append(
                        {
                            "op": "add",
                            "path": f"/spec/template/spec/containers/{i}/env",
                            "value": [{"name": "GOFIPS", "value": "1"}],
                        }
                    )

        if not patches:
            LOG.warning(f"No containers found in {resource_type}/{name}")
            continue

        # Apply the patch
        LOG.info(f"Patching {resource_type}/{name} to add GOFIPS=1")
        patch_json = json.dumps(patches)
        result = instance.exec(
            [
                "k8s",
                "kubectl",
                "patch",
                resource_type,
                name,
                "-n",
                namespace,
                "--type=json",
                "-p",
                patch_json,
            ],
            capture_output=True,
            check=False,
            text=True,
        )

        if result.returncode != 0:
            LOG.error(f"Failed to patch {resource_type}/{name}: {result.stderr}")
            continue

        LOG.info(f"Successfully patched {resource_type}/{name}")

        # Wait for pods to be recreated and potentially fail
        # We need to give time for:
        # 1. Old pods to be terminated
        # 2. New pods to be created with GOFIPS=1
        # 3. New pods to fail due to FIPS error
        LOG.info(f"Waiting for {resource_type}/{name} pods to restart with GOFIPS=1...")
        time.sleep(10)

        # Retry a few times to check for FIPS errors in pods
        found_fips_error = False
        for attempt in range(5):
            LOG.info(f"Attempt {attempt + 1} to find FIPS errors in pods")

            # Check that pods are failing with FIPS error
            # Get all pods in namespace and check their status
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

            # Check for pods that match this resource (by owner references or labels)
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
                        resource_type == "deployment"
                        and owner.get("kind") == "ReplicaSet"
                    ):
                        # Check if ReplicaSet is owned by our deployment
                        rs_name = owner.get("name")
                        if rs_name.startswith(name + "-"):
                            is_owned = True
                            break

                if not is_owned:
                    continue

                # Check pod events and logs for FIPS error
                LOG.info(f"Checking pod {pod_name} for FIPS errors")

                # Try to get logs (pods might be in CrashLoopBackOff)
                # First try current logs
                log_result = instance.exec(
                    [
                        "k8s",
                        "kubectl",
                        "logs",
                        pod_name,
                        "-n",
                        namespace,
                        "--tail=50",
                    ],
                    capture_output=True,
                    check=False,
                    text=True,
                )

                if log_result.returncode == 0:
                    logs = log_result.stdout
                    if (
                        "FIPS mode requested" in logs
                        or "FIPS mode" in logs
                        or "opensslcrypto" in logs
                    ):
                        LOG.info(f"Found FIPS error in logs for pod {pod_name}")
                        found_fips_error = True
                        break

                # Also try previous logs if container crashed
                log_result_prev = instance.exec(
                    [
                        "k8s",
                        "kubectl",
                        "logs",
                        pod_name,
                        "-n",
                        namespace,
                        "--previous",
                        "--tail=50",
                    ],
                    capture_output=True,
                    check=False,
                    text=True,
                )

                if log_result_prev.returncode == 0:
                    logs = log_result_prev.stdout
                    if (
                        "FIPS mode requested" in logs
                        or "FIPS mode" in logs
                        or "opensslcrypto" in logs
                    ):
                        LOG.info(
                            f"Found FIPS error in previous logs for pod {pod_name}"
                        )
                        found_fips_error = True
                        break

                # Also check pod status
                container_statuses = pod.get("status", {}).get("containerStatuses", [])
                for container_status in container_statuses:
                    waiting = container_status.get("state", {}).get("waiting", {})
                    terminated = container_status.get("state", {}).get("terminated", {})

                    reason = waiting.get("reason", "") or terminated.get("reason", "")
                    message = waiting.get("message", "") or terminated.get(
                        "message", ""
                    )

                    if "CrashLoopBackOff" in reason or "Error" in reason:
                        LOG.info(
                            f"Pod {pod_name} is in error state: {reason} - {message}"
                        )
                        if "FIPS" in message or "opensslcrypto" in message:
                            found_fips_error = True
                            break

            if found_fips_error:
                break

            # Wait before retrying
            if attempt < 4:
                LOG.info("No FIPS errors found yet, waiting 10 seconds...")
                time.sleep(10)

        assert found_fips_error, (
            f"Expected pods for {resource_type}/{name} in {namespace} to fail with FIPS error, "
            "but no FIPS-related errors were found"
        )

        LOG.info(f"Verified FIPS error for {resource_type}/{name}")

    LOG.info("All container images successfully verified as FIPS-compiled")
