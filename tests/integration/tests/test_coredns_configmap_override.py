#
# Copyright 2025 Canonical, Ltd.
#
"""Integration test for the CoreDNS ConfigMap override feature.

This test verifies that the k8sd CoreDNSConfigMapController picks up changes
to the `k8sd-coredns-values` ConfigMap in kube-system and applies them to
the CoreDNS Helm release (ck-dns).

Test flow:
1. Bootstrap a cluster and wait for DNS to be ready.
2. Apply an override ConfigMap with custom HPA replica counts.
3. Poll Helm values until the HPA override is reflected.
4. Update the ConfigMap (new HPA counts) and verify the new values are applied.
5. Update the ConfigMap with resource limits (a value k8sd does not set itself)
   and verify the limits appear in the Helm release values.
6. Delete the ConfigMap and verify:
   a. HPA values revert to the k8sd defaults (minReplicas=2, maxReplicas=100).
   b. The resource limits are absent from the Helm release values (never set by k8sd).
"""

import logging
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

# The ConfigMap name and namespace that k8sd watches for CoreDNS Helm overrides.
COREDNS_OVERRIDE_CM_NAME = "k8sd-coredns-values"
COREDNS_OVERRIDE_CM_NAMESPACE = "kube-system"

# The Helm release name for CoreDNS and the namespace it is deployed to.
COREDNS_HELM_RELEASE = "ck-dns"
COREDNS_HELM_NAMESPACE = "kube-system"


def _apply_coredns_override_configmap(instance: harness.Instance, values_yaml: str):
    """Create or replace the CoreDNS override ConfigMap with the given values YAML."""
    manifest = (
        f"apiVersion: v1\n"
        f"kind: ConfigMap\n"
        f"metadata:\n"
        f"  name: {COREDNS_OVERRIDE_CM_NAME}\n"
        f"  namespace: {COREDNS_OVERRIDE_CM_NAMESPACE}\n"
        f"data:\n"
        f"  values: |\n"
    )
    # Indent the values YAML as a block scalar under the `values` key.
    for line in values_yaml.splitlines():
        manifest += f"    {line}\n"

    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.encode(),
    )


def _delete_coredns_override_configmap(instance: harness.Instance):
    """Delete the CoreDNS override ConfigMap, ignoring not-found errors."""
    instance.exec(
        [
            "k8s",
            "kubectl",
            "delete",
            "configmap",
            COREDNS_OVERRIDE_CM_NAME,
            "-n",
            COREDNS_OVERRIDE_CM_NAMESPACE,
            "--ignore-not-found",
        ],
    )


def _helm_values_cmd() -> List[str]:
    return [
        "k8s",
        "helm",
        "get",
        "values",
        COREDNS_HELM_RELEASE,
        "--namespace",
        COREDNS_HELM_NAMESPACE,
        "--output",
        "yaml",
    ]


def _parse_helm_stdout(p) -> dict:
    """Parse YAML from a CompletedProcess whose stdout is bytes or str."""
    raw = p.stdout
    if isinstance(raw, bytes):
        raw = raw.decode()
    return yaml.safe_load(raw) or {}


def _wait_for_hpa_values(
    instance: harness.Instance,
    expected_min: int,
    expected_max: int,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until the HPA replica counts match the expected values."""

    def _hpa_matches(p) -> bool:
        values = _parse_helm_stdout(p)
        hpa = values.get("hpa", {})
        actual_min = hpa.get("minReplicas")
        actual_max = hpa.get("maxReplicas")
        LOG.info(
            "Helm HPA values \u2014 minReplicas: %s (want %s), maxReplicas: %s (want %s)",
            actual_min,
            expected_min,
            actual_max,
            expected_max,
        )
        return actual_min == expected_min and actual_max == expected_max

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _hpa_matches
    ).exec(
        _helm_values_cmd(),
    )


def _wait_for_resource_limits(
    instance: harness.Instance,
    expected_cpu: str,
    expected_memory: str,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until resource limits match the expected values.

    This exercises an override for a value that k8sd does NOT set itself,
    ensuring the ConfigMap override can inject arbitrary chart values.
    """

    def _resources_match(p) -> bool:
        values = _parse_helm_stdout(p)
        limits = values.get("resources", {}).get("limits", {})
        actual_cpu = limits.get("cpu")
        actual_memory = limits.get("memory")
        LOG.info(
            "Helm resource limits \u2014 cpu: %s (want %s), memory: %s (want %s)",
            actual_cpu,
            expected_cpu,
            actual_memory,
            expected_memory,
        )
        return actual_cpu == expected_cpu and actual_memory == expected_memory

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _resources_match
    ).exec(
        _helm_values_cmd(),
    )


def _wait_for_defaults_restored(instance: harness.Instance):
    """Poll until Helm values reflect k8sd defaults after ConfigMap deletion.

    After the ConfigMap is deleted the controller re-runs ApplyDNS with no
    overrides.  k8sd always passes an explicit `hpa` block so it will revert
    to the k8sd-defined defaults (minReplicas=2, maxReplicas=100).  The
    `resources` block was never set by k8sd, so it should be absent entirely.
    """

    def _defaults_restored(p) -> bool:
        values = _parse_helm_stdout(p)
        hpa = values.get("hpa", {})
        actual_min = hpa.get("minReplicas")
        actual_max = hpa.get("maxReplicas")
        resources_absent = "resources" not in values

        LOG.info(
            "Checking defaults \u2014 hpa.minReplicas: %s (want 2), "
            "hpa.maxReplicas: %s (want 100), resources absent: %s (want True)",
            actual_min,
            actual_max,
            resources_absent,
        )
        return actual_min == 2 and actual_max == 100 and resources_absent

    util.stubbornly(retries=30, delay_s=5).on(instance).until(_defaults_restored).exec(
        _helm_values_cmd(),
    )


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_coredns_configmap_override(instances: List[harness.Instance]):
    """Verify that the CoreDNSConfigMapController applies and reverts Helm overrides.

    Steps:
    1. Wait for the cluster and DNS to be ready.
    2. Apply a ConfigMap with HPA overrides (minReplicas=4, maxReplicas=60).
    3. Confirm the Helm release reflects the new values.
    4. Update the ConfigMap (minReplicas=6, maxReplicas=30).
    5. Confirm the Helm release reflects the updated values.
    6. Update the ConfigMap to also set resource limits \u2014 a value k8sd never
       passes itself \u2014 to verify arbitrary chart values can be injected.
    7. Confirm the resource limits appear in the Helm release values.
    8. Delete the ConfigMap.
    9. Confirm HPA reverts to k8sd defaults (minReplicas=2, maxReplicas=100)
       and the resource limits are absent from the Helm release values.
    """
    instance = instances[0]

    # Ensure cleanup even on test failure.
    try:
        util.wait_until_k8s_ready(instance, [instance])
        util.wait_for_dns(instance)

        # --- Step 1: Apply initial ConfigMap override ---
        LOG.info(
            "Applying CoreDNS override ConfigMap with minReplicas=4, maxReplicas=60"
        )
        _apply_coredns_override_configmap(
            instance,
            "hpa:\n  minReplicas: 4\n  maxReplicas: 60\n",
        )

        LOG.info("Waiting for Helm to reflect minReplicas=4, maxReplicas=60")
        _wait_for_hpa_values(instance, expected_min=4, expected_max=60)

        # --- Step 2: Update the ConfigMap override ---
        LOG.info("Updating CoreDNS override ConfigMap to minReplicas=6, maxReplicas=30")
        _apply_coredns_override_configmap(
            instance,
            "hpa:\n  minReplicas: 6\n  maxReplicas: 30\n",
        )

        LOG.info("Waiting for Helm to reflect minReplicas=6, maxReplicas=30")
        _wait_for_hpa_values(instance, expected_min=6, expected_max=30)

        # --- Step 3: Override a value k8sd does not set (resource limits) ---
        # k8sd never passes `resources` to the CoreDNS chart, so this tests
        # that the ConfigMap can inject completely new chart values.
        LOG.info(
            "Updating ConfigMap to also set resource limits "
            "(cpu=200m, memory=170Mi) \u2014 a value k8sd does not pass itself"
        )
        _apply_coredns_override_configmap(
            instance,
            "hpa:\n  minReplicas: 6\n  maxReplicas: 30\n"
            "resources:\n  limits:\n    cpu: 200m\n    memory: 170Mi\n",
        )

        LOG.info("Waiting for Helm to reflect resource limits cpu=200m, memory=170Mi")
        _wait_for_resource_limits(
            instance, expected_cpu="200m", expected_memory="170Mi"
        )

        # --- Step 4: Delete the ConfigMap and verify revert to defaults ---
        LOG.info("Deleting CoreDNS override ConfigMap")
        _delete_coredns_override_configmap(instance)

        LOG.info(
            "Waiting for HPA to revert to k8sd defaults (min=2, max=100) "
            "and resource limits to be absent from Helm values"
        )
        _wait_for_defaults_restored(instance)

    finally:
        # Best-effort cleanup so the override doesn't affect subsequent tests.
        _delete_coredns_override_configmap(instance)
