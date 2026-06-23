#
# Copyright 2026 Canonical, Ltd.
#
"""Shared utilities for ConfigMap override integration tests."""

import logging
from typing import List

import yaml
from test_util import harness, util

LOG = logging.getLogger(__name__)


def apply_override_configmap(
    instance: harness.Instance,
    cm_name: str,
    cm_namespace: str,
    values_yaml: str,
):
    """Apply a ConfigMap with Helm override values."""
    manifest = (
        f"apiVersion: v1\n"
        f"kind: ConfigMap\n"
        f"metadata:\n"
        f"  name: {cm_name}\n"
        f"  namespace: {cm_namespace}\n"
        f"data:\n"
        f"  values: |\n"
    )
    for line in values_yaml.splitlines():
        manifest += f"    {line}\n"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=manifest.encode(),
    )


def delete_override_configmap(
    instance: harness.Instance,
    cm_name: str,
    cm_namespace: str,
):
    """Delete a ConfigMap override."""
    instance.exec(
        [
            "k8s",
            "kubectl",
            "delete",
            "configmap",
            cm_name,
            "-n",
            cm_namespace,
            "--ignore-not-found",
        ],
    )


def helm_values_cmd(helm_release: str, helm_namespace: str) -> List[str]:
    """Return the command to get Helm release values."""
    return [
        "k8s",
        "helm",
        "get",
        "values",
        helm_release,
        "--namespace",
        helm_namespace,
        "--output",
        "yaml",
    ]


def parse_helm_stdout(p) -> dict:
    """Parse YAML output from helm command."""
    raw = p.stdout
    if isinstance(raw, bytes):
        raw = raw.decode()
    return yaml.safe_load(raw) or {}


def wait_for_override(
    instance: harness.Instance,
    helm_release: str,
    helm_namespace: str,
    expected_key_path: List[str],
    expected_value,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until the nested key matches the expected value."""

    def _value_matches(p) -> bool:
        values = parse_helm_stdout(p)
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
        helm_values_cmd(helm_release, helm_namespace),
    )

