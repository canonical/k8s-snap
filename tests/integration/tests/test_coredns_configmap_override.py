#
# Copyright 2025 Canonical, Ltd.
#
"""Integration test for the CoreDNS ConfigMap override feature.

This test verifies that the k8sd CoreDNSConfigMapController picks up changes
to the `k8sd-coredns-values` ConfigMap in kube-system and applies them to
the CoreDNS Helm release (ck-dns).

Test flow:
1. Bootstrap a cluster and wait for DNS to be ready.
2. Apply the override ConfigMap with custom HPA replica counts.
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap and verify the new values are applied.
5. Delete the ConfigMap and verify Helm values revert to defaults.
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
        "/snap/k8s/current/bin/helm",
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
            "Helm HPA values — minReplicas: %s (want %s), maxReplicas: %s (want %s)",
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
        env={"KUBECONFIG": "/etc/kubernetes/admin.conf"},
    )


def _wait_for_default_hpa_values(instance: harness.Instance):
    """Poll until Helm values no longer contain user-supplied HPA overrides.

    When the ConfigMap is deleted, the controller re-runs ApplyDNS with no
    overrides, so the Helm release reverts to chart defaults.  The resulting
    `helm get values` output will either be null/empty or will no longer
    contain the HPA block we injected.
    """

    def _defaults_restored(p) -> bool:
        values = _parse_helm_stdout(p)
        hpa = values.get("hpa", {})
        if not hpa:
            LOG.info("HPA overrides removed from Helm values — defaults restored")
            return True
        LOG.info("HPA still in Helm values: %s", hpa)
        return False

    util.stubbornly(retries=30, delay_s=5).on(instance).until(
        _defaults_restored
    ).exec(
        _helm_values_cmd(),
        env={"KUBECONFIG": "/etc/kubernetes/admin.conf"},
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
    6. Delete the ConfigMap.
    7. Confirm the HPA overrides are gone from the Helm release.
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
        LOG.info(
            "Updating CoreDNS override ConfigMap to minReplicas=6, maxReplicas=30"
        )
        _apply_coredns_override_configmap(
            instance,
            "hpa:\n  minReplicas: 6\n  maxReplicas: 30\n",
        )

        LOG.info("Waiting for Helm to reflect minReplicas=6, maxReplicas=30")
        _wait_for_hpa_values(instance, expected_min=6, expected_max=30)

        # --- Step 3: Delete the ConfigMap and verify revert to defaults ---
        LOG.info("Deleting CoreDNS override ConfigMap")
        _delete_coredns_override_configmap(instance)

        LOG.info("Waiting for HPA overrides to be removed from Helm values")
        _wait_for_default_hpa_values(instance)

    finally:
        # Best-effort cleanup so the override doesn't affect subsequent tests.
        _delete_coredns_override_configmap(instance)
