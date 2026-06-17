#
# Copyright 2025 Canonical, Ltd.
#
"""Integration test for Cilium network interface selection via ConfigMap override.

Product feedback context:
  Canonical K8s does not expose a configuration option for selecting which
  network interface Cilium should use for VxLAN pod-to-pod communications.
  On bare-metal nodes with multiple interfaces (e.g., 1 Gbps management +
  25/100 Gbps data-plane), Cilium may route tunnel traffic over the slow
  management interface, causing performance issues in production.

This test verifies that an infrastructure administrator can address this
problem as a day-2 operation by creating the ``k8sd-cilium-values``
ConfigMap in kube-system with the ``devices`` Helm value, without
restarting or re-bootstrapping the cluster.

Test flow:
1. Bootstrap a cluster and wait for network to be ready.
2. Discover an available network interface on the node.
3. Apply the override ConfigMap with ``devices: <interface>``.
4. Poll ``helm get values ck-network`` until the override is reflected.
5. Update the ConfigMap to a different interface name and verify the
   new value is applied (simulates day-2 reconfiguration).
6. Delete the ConfigMap and verify the ``devices`` key is absent from
   the Helm release values (Cilium reverts to automatic detection).
"""

import logging
from typing import List

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-cilium-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "ck-network"
HELM_NAMESPACE = "kube-system"


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
        "k8s",
        "helm",
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


def _detect_network_interface(instance: harness.Instance) -> str:
    """Return the name of the first non-loopback network interface on the node.

    This simulates the admin identifying the desired data-plane interface.
    On a real bare-metal server this might be ``eth1`` or ``ens3``.
    """
    p = instance.exec(
        ["ip", "-o", "link", "show"],
        capture_output=True,
    )
    lines = p.stdout.decode() if isinstance(p.stdout, bytes) else p.stdout
    for line in lines.splitlines():
        # ip -o link output: "<idx>: <name>: <flags>"
        parts = line.split(":")
        if len(parts) < 2:
            continue
        name = parts[1].strip().split("@")[0]
        if name == "lo":
            continue
        LOG.info("Detected network interface: %s", name)
        return name
    raise RuntimeError("No non-loopback network interface found on test node")


def _wait_for_devices(
    instance: harness.Instance,
    expected_devices: str,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until devices matches the expected value."""

    def _devices_match(p) -> bool:
        values = _parse_helm_stdout(p)
        actual = values.get("devices")
        LOG.info("Helm devices: %r (want %r)", actual, expected_devices)
        return actual == expected_devices

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _devices_match
    ).exec(
        _helm_values_cmd(),
    )


def _wait_for_devices_absent(
    instance: harness.Instance,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until 'devices' is absent (Cilium reverts to auto-detection)."""

    def _devices_absent(p) -> bool:
        values = _parse_helm_stdout(p)
        absent = "devices" not in values
        LOG.info("Helm 'devices' absent: %s (want True)", absent)
        return absent

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _devices_absent
    ).exec(
        _helm_values_cmd(),
    )


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_cilium_devices_configmap_override(instances: List[harness.Instance]):
    """Day-2 Cilium interface selection via ConfigMap override.

    Verifies the solution to the product feedback: an admin can specify
    which network interface Cilium uses for VxLAN pod-to-pod communications
    after the cluster is already running, without re-bootstrapping.

    Steps:
    1. Detect a network interface available on the node.
    2. Apply override ConfigMap with ``devices: <interface>``.
    3. Verify Helm release reflects the new value (day-2 configuration applied).
    4. Update the override to a wildcard pattern (``eth+``) simulating a
       real-world scenario where multiple fast interfaces should be selected.
    5. Verify the updated value is applied.
    6. Delete the ConfigMap and verify Cilium reverts to automatic detection
       (``devices`` key absent from Helm values).
    """
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # Discover an available interface on the node to use as the override target.
        iface = _detect_network_interface(instance)

        # --- Step 1: Apply initial override (specific interface) ---
        # This simulates an admin pinning Cilium to a specific fast interface
        # on a bare-metal node with multiple NICs (e.g., the 25/100 Gbps data-plane
        # interface, not the 1 Gbps management interface).
        LOG.info(
            "Applying Cilium override ConfigMap: devices=%s (day-2 interface selection)",
            iface,
        )
        _apply_override_configmap(instance, f"devices: {iface}\n")

        LOG.info("Waiting for Helm to reflect devices=%s", iface)
        _wait_for_devices(instance, expected_devices=iface)

        # --- Step 2: Update to wildcard pattern ---
        # Simulate updating configuration on an already-running cluster
        # (day-2 reconfiguration \u2014 key requirement from product feedback).
        LOG.info("Updating Cilium override ConfigMap: devices=eth+")
        _apply_override_configmap(instance, "devices: eth+\n")

        LOG.info("Waiting for Helm to reflect devices=eth+")
        _wait_for_devices(instance, expected_devices="eth+")

        # --- Step 3: Delete override, verify revert to automatic detection ---
        LOG.info("Deleting Cilium override ConfigMap")
        _delete_override_configmap(instance)

        LOG.info(
            "Waiting for 'devices' to be absent from Helm values "
            "(Cilium reverts to automatic interface detection)"
        )
        _wait_for_devices_absent(instance)

    finally:
        _delete_override_configmap(instance)
