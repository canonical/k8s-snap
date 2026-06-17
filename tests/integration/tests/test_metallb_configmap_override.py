#
# Copyright 2025 Canonical, Ltd.
#
"""Integration test for the MetalLB ConfigMap override feature.

This test verifies that the k8sd MetalLBConfigMapController picks up changes
to the `k8sd-metallb-values` ConfigMap in kube-system and applies them to
the MetalLB Helm release (metallb in metallb-system).

Test flow:
1. Bootstrap a cluster with load-balancer enabled, wait for it to be ready.
2. Apply an override ConfigMap with a value k8sd does not set.
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap and verify the new value is applied.
5. Delete the ConfigMap and verify the override is absent from Helm values.
"""

import logging
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-metallb-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "metallb"
HELM_NAMESPACE = "metallb-system"


def _apply_override_configmap(instance: harness.Instance, values_yaml: str):
    manifest = (
        f"apiVersion: v1\n"
        f"kind: ConfigMap\n"
        f"metadata:\n"
        f"  name: {OVERRIDE_CM_NAME}\n"
        f"  namespace: {OVERRIDE_CM_NAMESPACE}\n"
        f"data:\n"
        f"  values: |\n"
    )
    for line in values_yaml.splitlines():
        manifest += f"    {line}\n"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.encode(),
    )


def _delete_override_configmap(instance: harness.Instance):
    instance.exec(
        [
            "k8s",
            "kubectl",
            "delete",
            "configmap",
            OVERRIDE_CM_NAME,
            "-n",
            OVERRIDE_CM_NAMESPACE,
            "--ignore-not-found",
        ],
    )


def _helm_values_cmd() -> List[str]:
    return [
        "/snap/k8s/current/bin/helm",
        "get",
        "values",
        HELM_RELEASE,
        "--namespace",
        HELM_NAMESPACE,
        "--output",
        "yaml",
    ]


def _parse_helm_stdout(p) -> dict:
    raw = p.stdout
    if isinstance(raw, bytes):
        raw = raw.decode()
    return yaml.safe_load(raw) or {}


def _wait_for_override(
    instance: harness.Instance,
    expected_key_path: List[str],
    expected_value,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until the nested key matches the expected value."""

    def _value_matches(p) -> bool:
        values = _parse_helm_stdout(p)
        node = values
        for key in expected_key_path:
            if not isinstance(node, dict) or key not in node:
                LOG.info(
                    "Key path %s not yet present in Helm values", expected_key_path
                )
                return False
            node = node[key]
        match = node == expected_value
        LOG.info(
            "Helm value at %s: %s (want %s)",
            ".".join(expected_key_path),
            node,
            expected_value,
        )
        return match

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _value_matches
    ).exec(
        _helm_values_cmd(),
        env={"KUBECONFIG": "/etc/kubernetes/admin.conf"},
    )


def _wait_for_key_absent(
    instance: harness.Instance,
    top_level_key: str,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until the top-level key is absent."""

    def _key_absent(p) -> bool:
        values = _parse_helm_stdout(p)
        absent = top_level_key not in values
        LOG.info("Helm values key '%s' absent: %s (want True)", top_level_key, absent)
        return absent

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _key_absent
    ).exec(
        _helm_values_cmd(),
        env={"KUBECONFIG": "/etc/kubernetes/admin.conf"},
    )


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_metallb_configmap_override(instances: List[harness.Instance]):
    """Verify that the MetalLBConfigMapController applies and reverts Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # --- Step 1: Apply initial ConfigMap override ---
        # Override controller.logLevel (k8sd does not set this value, so it will
        # appear in helm get values only when explicitly overridden).
        LOG.info("Applying MetalLB override ConfigMap with controller.logLevel=debug")
        _apply_override_configmap(
            instance,
            "controller:\n  logLevel: debug\n",
        )

        LOG.info("Waiting for Helm to reflect controller.logLevel=debug")
        _wait_for_override(instance, ["controller", "logLevel"], "debug")

        # --- Step 2: Update the ConfigMap override ---
        LOG.info("Updating MetalLB override ConfigMap with controller.logLevel=info")
        _apply_override_configmap(
            instance,
            "controller:\n  logLevel: info\n",
        )

        LOG.info("Waiting for Helm to reflect controller.logLevel=info")
        _wait_for_override(instance, ["controller", "logLevel"], "info")

        # --- Step 3: Delete the ConfigMap and verify revert ---
        LOG.info("Deleting MetalLB override ConfigMap")
        _delete_override_configmap(instance)

        LOG.info("Waiting for controller key to be absent from Helm values")
        _wait_for_key_absent(instance, "controller")

    finally:
        _delete_override_configmap(instance)
