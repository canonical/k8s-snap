#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from dataclasses import dataclass
from fnmatch import fnmatch
from typing import Any, Dict, List, Optional

from test_util import harness

LOG = logging.getLogger(__name__)


@dataclass
class Resource:
    namespace: str
    type: str
    name: str


def get_workload_resources_in_namespaces(
    instance: harness.Instance,
    namespaces: List[str],
    workload_types: List[str],
    exclude: Optional[List[str]] = None,
) -> List[Resource]:
    """
    Get workload resources (apps-style resources with pod templates) of specified types
    in specified namespaces.

    This function is intended for resources that have a pod template containing
    containers (for example: "deployment", "daemonset", "statefulset"). It is not
    suitable for non-workload kinds like Service or Ingress.

    Args:
        instance: instance on which to execute kubectl
        namespaces: list of namespace names to check
        workload_types: list of workload types (e.g., "deployment", "daemonset")
        exclude: optional list of name patterns to exclude (supports wildcards, e.g. "*ck-storage*")

    Returns:
        list of Resource objects with namespace, type, and name
    """
    resources = []
    exclude = exclude or []

    for namespace in namespaces:
        for workload_type in workload_types:
            LOG.info(f"Checking workload {workload_type}s in namespace {namespace}")
            result = instance.exec(
                [
                    "k8s",
                    "kubectl",
                    "get",
                    workload_type,
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
                    f"No {workload_type}s found in namespace {namespace} or namespace doesn't exist"
                )
                continue

            resources_data = json.loads(result.stdout)

            for item in resources_data.get("items", []):
                name = item["metadata"]["name"]

                # Be defensive when accessing nested spec/template/spec/containers
                template = item.get("spec", {}).get("template", {})
                pod_spec = template.get("spec", {})
                containers = pod_spec.get("containers", [])

                # Exclude resources whose name matches any of the exclude patterns
                excluded = False
                for pattern in exclude:
                    if fnmatch(name, pattern):
                        LOG.info(
                            f"Excluding workload {workload_type}/{name} in {namespace} (pattern: {pattern})"
                        )
                        excluded = True
                        break
                if excluded:
                    continue

                if containers:
                    resources.append(
                        Resource(
                            namespace=namespace,
                            type=workload_type,
                            name=name,
                        )
                    )
                    LOG.info(f"Found workload: {workload_type}/{name} in {namespace}")

    return resources


def update_resource_container_env(
    instance: harness.Instance,
    namespace: str,
    resource_type: str,
    name: str,
    env_vars: Dict[str, str],
    containers: Optional[List[str]] = None,
):
    """
    Update or add environment variables for a resource's containers.

    Args:
        instance: instance used to run the kubectl commands
        namespace: namespace of the resource
        resource_type: kind of resource (for example "deployment" or "daemonset")
        name: name of the resource
        env_vars: mapping of environment variable names to values to set
        containers: optional list of container names to update; if None, all containers are updated

    Raises:
        RuntimeError: if the kubectl patch command fails
    """
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

    containers_defs = resource_def["spec"]["template"]["spec"]["containers"]
    patches: List[Dict[str, Any]] = []

    for i, container in enumerate(containers_defs):
        if containers and container["name"] not in containers:
            logging.info(
                f"Skipping container {container['name']} not in specified containers list ({containers})"
            )
            continue

        env = container.get("env", [])
        existing_env_names = {e.get("name"): j for j, e in enumerate(env)}

        for key, val in env_vars.items():
            if key in existing_env_names:
                j = existing_env_names[key]
                patches.append(
                    {
                        "op": "replace",
                        "path": f"/spec/template/spec/containers/{i}/env/{j}/value",
                        "value": val,
                    }
                )
            else:
                if env:
                    patches.append(
                        {
                            "op": "add",
                            "path": f"/spec/template/spec/containers/{i}/env/-",
                            "value": {"name": key, "value": val},
                        }
                    )
                else:
                    patches.append(
                        {
                            "op": "add",
                            "path": f"/spec/template/spec/containers/{i}/env",
                            "value": [{"name": key, "value": val}],
                        }
                    )

    if not patches:
        raise ValueError(f"No containers found in {resource_type}/{name}")

    # Apply the patch
    LOG.info(f"Patching {resource_type}/{name} to modify {env_vars}")
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
        raise RuntimeError(f"Failed to patch {resource_type}/{name}")

    LOG.info(f"Successfully patched {resource_type}/{name}")


def resource_ready(
    instance: harness.Instance,
    namespace: str,
    resource_type: str,
    name: str,
) -> bool:
    """
    Check if a k8s resource is in ready state.

    Args:
        instance: instance on which to execute check
        namespace: namespace of the resource
        resource_type: type of resource (e.g., "deployment", "daemonset")
        name: name of the resource
    Returns:
        True if resource is ready, False otherwise
    """
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
        check=False,
        text=True,
    )

    if result.returncode != 0:
        LOG.error(f"Failed to get {resource_type}/{name} in {namespace}")
        return False

    resource_def = json.loads(result.stdout)

    if resource_type == "deployment":
        desired = resource_def["status"].get("replicas", 0)
        available = resource_def["status"].get("availableReplicas", 0)
        return desired == available
    elif resource_type == "daemonset":
        desired = resource_def["status"].get("desiredNumberScheduled", 0)
        available = resource_def["status"].get("numberReady", 0)
        return desired == available
    elif resource_type == "statefulset":
        desired = resource_def["status"].get("replicas", 0)
        ready = resource_def["status"].get("readyReplicas", 0)
        return desired == ready
    elif resource_type == "pod":
        phase = resource_def["status"].get("phase", "")
        return phase == "Running"
    else:
        LOG.error(f"Unsupported resource type: {resource_type}")
        return False
