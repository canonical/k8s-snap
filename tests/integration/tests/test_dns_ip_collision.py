#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import re
import subprocess
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

DUMMY_SERVICE_YAML = """apiVersion: v1
kind: Service
metadata:
  name: {name}
  namespace: {namespace}
spec:
  clusterIP: {ip}
  ports:
  - port: 80
    targetPort: 80
"""


@pytest.mark.tags(tags.PULL_REQUEST)
def test_dns_auto_ip_reallocated_after_collision(instances: List[harness.Instance]):
    """
    Validates that if an auto-assigned CoreDNS ClusterIP is occupied by another service
    while DNS is disabled, enabling DNS recovers by dynamically allocating a new ClusterIP,
    updating cluster-dns in node configuration, and restoring in-cluster DNS resolution.
    (Regression test for canonical/k8sd#227).
    """
    instance = instances[0]

    LOG.info("Waiting for cluster to be ready with DNS enabled")
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_dns(instance)

    # 1. Record the initial auto-assigned CoreDNS ClusterIP
    result = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "kube-system",
            "coredns",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        capture_output=True,
        text=True,
    )
    old_dns_ip = result.stdout.strip()
    assert old_dns_ip, "Expected initial CoreDNS service to have a ClusterIP"
    LOG.info("Initial CoreDNS ClusterIP: %s", old_dns_ip)

    # Verify k8s get dns.service-ip matches initial IP
    res = instance.exec(
        ["k8s", "get", "dns.service-ip"], capture_output=True, text=True
    )
    assert (
        res.stdout.strip() == old_dns_ip
    ), f"Expected dns.service-ip to be {old_dns_ip}"

    # 2. Disable DNS and wait until the CoreDNS service is gone
    LOG.info("Disabling DNS")
    instance.exec(["k8s", "disable", "dns"])

    LOG.info("Waiting for CoreDNS service to be deleted")
    util.stubbornly(retries=30, delay_s=2).on(instance).until(
        lambda p: p.returncode != 0
    ).exec(
        ["k8s", "kubectl", "get", "service", "-n", "kube-system", "coredns"],
        check=False,
    )

    # 3. Create dummy service occupying the old CoreDNS IP
    LOG.info("Creating dummy service to occupy old DNS IP %s", old_dns_ip)
    dummy_yaml = DUMMY_SERVICE_YAML.format(
        name="dummy-collision",
        namespace="default",
        ip=old_dns_ip,
    )
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=dummy_yaml.encode(),
    )

    res = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "default",
            "dummy-collision",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        capture_output=True,
        text=True,
    )
    assert (
        res.stdout.strip() == old_dns_ip
    ), "Dummy service failed to allocate old DNS IP"

    # 4. Enable DNS
    LOG.info("Re-enabling DNS; expecting dynamic reallocation after collision")
    instance.exec(["k8s", "enable", "dns"])

    # 5. Assert that DNS comes back enabled with a DIFFERENT IP
    def dns_enabled_with_different_ip(p: subprocess.CompletedProcess) -> bool:
        try:
            status = json.loads(p.stdout)
            dns_status = status.get("dns", {})
            if not dns_status.get("enabled"):
                return False
            msg = dns_status.get("message", "")
            match = re.search(r"enabled at (\d{1,3}(?:\.\d{1,3}){3})", msg)
            if not match:
                return False
            return match.group(1) != old_dns_ip
        except Exception:
            return False

    LOG.info("Waiting for DNS to be reported enabled at a new IP")
    util.stubbornly(retries=40, delay_s=5).on(instance).until(
        dns_enabled_with_different_ip
    ).exec(
        ["k8s", "status", "--output-format=json"],
        text=True,
    )

    # Query new IP from CoreDNS service
    res = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "kube-system",
            "coredns",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        capture_output=True,
        text=True,
    )
    new_dns_ip = res.stdout.strip()
    assert new_dns_ip, "Expected new CoreDNS service to have a ClusterIP"
    assert (
        new_dns_ip != old_dns_ip
    ), f"DNS IP should have changed from {old_dns_ip}, but got {new_dns_ip}"
    LOG.info("New CoreDNS ClusterIP: %s", new_dns_ip)

    # Assert k8s get dns.service-ip reflects new IP
    res = instance.exec(
        ["k8s", "get", "dns.service-ip"], capture_output=True, text=True
    )
    assert (
        res.stdout.strip() == new_dns_ip
    ), f"Expected dns.service-ip to be {new_dns_ip}, got {res.stdout.strip()}"

    # Assert kubelet cluster-dns args reflect the new IP
    expected_arg = f'--cluster-dns="{new_dns_ip}"'
    LOG.info("Asserting kubelet args updated with %s", expected_arg)
    util.stubbornly(retries=20, delay_s=3).on(instance).until(
        lambda p: expected_arg
        in (p.stdout.decode() if isinstance(p.stdout, bytes) else p.stdout)
    ).exec(["cat", "/var/snap/k8s/common/args/kubelet"])

    # Wait for DNS pods to be ready
    util.wait_for_dns(instance)

    # 6. Assert in-cluster DNS resolution works from a pod
    LOG.info("Running busybox pod to verify in-cluster DNS resolution")
    instance.exec(
        [
            "k8s",
            "kubectl",
            "run",
            "busybox-dns-collision",
            "--image=ghcr.io/containerd/busybox:1.28",
            "--restart=Never",
            "--",
            "sleep",
            "3600",
        ]
    )

    util.stubbornly(retries=15, delay_s=2).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "busybox-dns-collision",
            "--timeout",
            "180s",
        ]
    )

    # Verify pod's /etc/resolv.conf points to the new DNS IP
    resolv = instance.exec(
        [
            "k8s",
            "kubectl",
            "exec",
            "busybox-dns-collision",
            "--",
            "cat",
            "/etc/resolv.conf",
        ],
        capture_output=True,
        text=True,
    ).stdout
    assert (
        f"nameserver {new_dns_ip}" in resolv
    ), f"Expected nameserver {new_dns_ip} in resolv.conf, got: {resolv}"

    def nslookup_matches(p: subprocess.CompletedProcess) -> bool:
        out = p.stdout.decode() if isinstance(p.stdout, bytes) else p.stdout
        return "10.152.183.1 kubernetes.default.svc.cluster.local" in out

    LOG.info("Verifying nslookup kubernetes.default from busybox pod")
    util.stubbornly(retries=15, delay_s=2).on(instance).until(nslookup_matches).exec(
        [
            "k8s",
            "kubectl",
            "exec",
            "busybox-dns-collision",
            "--",
            "nslookup",
            "kubernetes.default",
        ],
    )


@pytest.mark.tags(tags.PULL_REQUEST)
def test_dns_user_ip_collision_not_overridden(instances: List[harness.Instance]):
    """
    Validates that when a user explicitly configures dns.service-ip, an IP collision
    is NOT silently overridden by dynamic allocation. CoreDNS deployment fails and
    reports the allocation error, and the configured IP remains unchanged.
    (Regression test for canonical/k8sd#227).
    """
    instance = instances[0]

    LOG.info("Waiting for cluster to be ready")
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_dns(instance)

    # 1. Discover current DNS IP to determine service CIDR prefix
    result = instance.exec(
        ["k8s", "get", "dns.service-ip"],
        capture_output=True,
        text=True,
    )
    current_dns_ip = result.stdout.strip()
    assert current_dns_ip, "Expected initial DNS IP"
    ip_parts = current_dns_ip.split(".")
    target_host = "54" if ip_parts[3] == "53" else "53"
    user_dns_ip = f"{ip_parts[0]}.{ip_parts[1]}.{ip_parts[2]}.{target_host}"
    LOG.info("Selected user static DNS IP in service CIDR: %s", user_dns_ip)

    # 2. Disable DNS before setting a custom service-ip (changing cluster-dns while enabled is disallowed)
    LOG.info("Disabling DNS")
    instance.exec(["k8s", "disable", "dns"])

    LOG.info("Waiting for CoreDNS service to be deleted")
    util.stubbornly(retries=30, delay_s=2).on(instance).until(
        lambda p: p.returncode != 0
    ).exec(
        ["k8s", "kubectl", "get", "service", "-n", "kube-system", "coredns"],
        check=False,
    )

    # 3. Explicitly configure dns.service-ip
    LOG.info("Setting user-configured dns.service-ip=%s", user_dns_ip)
    instance.exec(["k8s", "set", f"dns.service-ip={user_dns_ip}"])

    res = instance.exec(
        ["k8s", "get", "dns.service-ip"], capture_output=True, text=True
    )
    assert (
        res.stdout.strip() == user_dns_ip
    ), f"Expected dns.service-ip to be {user_dns_ip}"

    # 4. Occupy that static IP with a dummy Service
    LOG.info("Creating dummy service to occupy user DNS IP %s", user_dns_ip)
    dummy_yaml = DUMMY_SERVICE_YAML.format(
        name="dummy-user-collision",
        namespace="default",
        ip=user_dns_ip,
    )
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=dummy_yaml.encode(),
    )

    res = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "default",
            "dummy-user-collision",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        capture_output=True,
        text=True,
    )
    assert (
        res.stdout.strip() == user_dns_ip
    ), "Dummy service failed to allocate user DNS IP"

    # 5. Enable DNS
    LOG.info("Enabling DNS; expecting collision error without fallback override")
    instance.exec(["k8s", "enable", "dns"])

    # 6. Assert that DNS does NOT silently move to another IP:
    # Feature status must show the allocation failure.
    def dns_shows_allocation_failure(p: subprocess.CompletedProcess) -> bool:
        try:
            status = json.loads(p.stdout)
            dns_status = status.get("dns", {})
            if dns_status.get("enabled") is True:
                return False
            msg = dns_status.get("message", "")
            return (
                f"{user_dns_ip} is already allocated to Service "
                "default/dummy-user-collision" in msg
            )
        except Exception:
            return False

    LOG.info("Waiting for DNS feature status to reflect allocation failure")
    util.stubbornly(retries=30, delay_s=3).on(instance).until(
        dns_shows_allocation_failure
    ).exec(
        ["k8s", "status", "--output-format=json"],
        text=True,
    )

    # 7. Assert that the configured IP is unchanged
    res = instance.exec(
        ["k8s", "get", "dns.service-ip"], capture_output=True, text=True
    )
    assert (
        res.stdout.strip() == user_dns_ip
    ), f"Configured IP should remain {user_dns_ip}, got {res.stdout.strip()}"

    # 8. Assert that no CoreDNS service was created with a different IP
    res = instance.exec(
        [
            "k8s",
            "kubectl",
            "get",
            "service",
            "-n",
            "kube-system",
            "coredns",
            "-o",
            "jsonpath={.spec.clusterIP}",
        ],
        check=False,
        capture_output=True,
        text=True,
    )
    if res.returncode == 0 and res.stdout.strip():
        assert (
            res.stdout.strip() == user_dns_ip
        ), f"CoreDNS service IP unexpectedly moved to {res.stdout.strip()}"
